package hivemetastore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"

	"github.com/beltran/gohive"
	"github.com/beltran/gohive/hive_metastore"
	dataextractor "github.com/ubix/dripper/manager/repositories/data-extractor"
)

type config struct {
	metastoreHost   string
	metastorePort   int
	sparkHost       string
	sparkPort       int
	dataspaceHost   string
	dataserviceHost string
}

type hivemetastore struct {
	config                 config
	metastoreConfiguration *gohive.MetastoreConnectConfiguration
	sparkConfiguration     *gohive.ConnectConfiguration
	httpClient             *httpclient.Client
}

func (m *hivemetastore) getMetastoreConnection() (*gohive.HiveMetastoreClient, error) {
	conn, err := gohive.ConnectToMetastore(m.config.metastoreHost, m.config.metastorePort, "NOSASL", m.metastoreConfiguration)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (m *hivemetastore) getSparkConnection() (*gohive.Connection, error) {
	conn, err := gohive.Connect(m.config.sparkHost, m.config.sparkPort, "NONE", m.sparkConfiguration)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func NewHiveMetastore(metastoreHost string, metastorePort int, sparkHost string, sparkPort int, dataspaceHost, dataserviceHost string) (*hivemetastore, error) {
	config := config{
		metastoreHost,
		metastorePort,
		sparkHost,
		sparkPort,
		dataspaceHost,
		dataserviceHost,
	}

	metastoreConfiguration := gohive.NewMetastoreConnectConfiguration()
	metastoreConfiguration.Username = "spark" // Scratch dockerfile fix

	sparkConfiguration := gohive.NewConnectConfiguration()
	sparkConfiguration.Username = "spark" // Scratch dockerfile fix

	httpClient := httpclient.NewClient(
		httpclient.WithHTTPTimeout(1 * time.Minute),
	)

	client := &hivemetastore{config, metastoreConfiguration, sparkConfiguration, httpClient}

	// check if metastore is ok as health check
	metastoreCheck, err := client.getMetastoreConnection()
	if err != nil {
		return nil, err
	}

	metastoreCheck.Close()

	// check if spark is ok as health check
	sparkCheck, err := client.getSparkConnection()
	if err != nil {
		return nil, err
	}

	sparkCheck.Close()

	return client, nil
}

type MetastoreError struct {
	message string
}

func NewMetastoreError(message string) *MetastoreError {
	return &MetastoreError{message}
}

func (e *MetastoreError) Error() string {
	return fmt.Sprintf("Metastore Operation failed due: %s", e.message)
}

func (m *hivemetastore) CreateDatabase(ctx context.Context, connector *dataextractor.Connector) error {
	locationUri, err := getLocationUriFromConnector(connector)
	if err != nil {
		return err
	}

	conn, err := m.getMetastoreConnection()
	if err != nil {
		return err
	}

	defer conn.Close()

	dbName := connector.GetWarehouseDatabaseName()
	database := hive_metastore.Database{
		Name:        dbName,
		LocationUri: *locationUri,
	}
	err = conn.Client.CreateDatabase(ctx, &database)
	if err != nil {
		return err
	}

	connector.WarehouseDatabase = database.Name

	return err
}

func (m *hivemetastore) DeleteDatabase(ctx context.Context, connector *dataextractor.Connector) error {
	conn, err := m.getMetastoreConnection()
	if err != nil {
		return err
	}

	defer conn.Close()

	dbName := connector.GetWarehouseDatabaseName()

	conn.Client.DropDatabase(ctx, dbName, true, true)

	return err
}

func (m *hivemetastore) UpdateTables(ctx context.Context, connector *dataextractor.Connector) error {
	conn, err := m.getMetastoreConnection()
	if err != nil {
		return err
	}

	defer conn.Close()

	dbLocationUri, err := getLocationUriFromConnector(connector)
	if err != nil {
		return err
	}

	for _, table := range connector.Tables {
		tableHiveName := table.GetHiveName()
		hiveTable, err := conn.Client.GetTable(ctx, connector.WarehouseDatabase, tableHiveName)
		if err != nil {
			hiveTable = nil
		}

		if table.SyncEnabled {
			// table.Name is the JsonSchema table name which is the correct name in which airbyte writes the data
			tableFolderName := strings.ReplaceAll(table.Name, " ", "_")
			tableLocation := *dbLocationUri + "/" + tableFolderName + "/"

			serDeInfo := getDefaultSerDeInfo()

			columns := GetColumnsFromJsonSchema(table.JsonSchema)

			storageDescriptor := hive_metastore.StorageDescriptor{
				Cols:         columns,
				Location:     tableLocation,
				InputFormat:  parquetInputFormat,
				OutputFormat: parquetOutputFormat,
				SerdeInfo:    serDeInfo,
			}

			partitionKeys := getDefaultPartitionKeys()

			newTable := hive_metastore.Table{
				TableName:     tableHiveName,
				DbName:        connector.WarehouseDatabase,
				Owner:         defaultTableOwner,
				Sd:            &storageDescriptor,
				PartitionKeys: partitionKeys,
				TableType:     "EXTERNAL_TABLE",
				Temporary:     false,
				Parameters:    map[string]string{},
			}

			if hiveTable == nil {
				err = conn.Client.CreateTable(ctx, &newTable)
			} else {
				// Keep columns order on new table
				colToOrderMap := map[string]int{}
				for i, col := range hiveTable.Sd.Cols {
					colToOrderMap[col.Name] = i
				}
				sort.Slice(newTable.Sd.Cols, func(i, j int) bool {
					return colToOrderMap[newTable.Sd.Cols[i].Name] < colToOrderMap[newTable.Sd.Cols[j].Name]
				})

				err = conn.Client.AlterTable(ctx, connector.WarehouseDatabase, tableHiveName, &newTable)
			}
		} else {
			if hiveTable != nil {
				err = conn.Client.DropTable(ctx, connector.WarehouseDatabase, tableHiveName, false)
			} else {
				err = nil
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

type dataspaceElement struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
}

type dataspaceDataset struct {
	WorkspaceId   string             `json:"workspaceId"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	Database      string             `json:"database"`
	Creator       string             `json:"create"`
	Frequency     string             `json:"frequency"`
	FrequencyCron string             `json:"frequencyCron"`
	StartDate     string             `json:"startDate"`
	DatsetType    string             `json:"datsetType"`
	Elements      []dataspaceElement `json:"elements"`
}

type dataspaceApiResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Success bool        `json:"success"`
}

// TODO: this will be replaced by the dripper interface api
func (m *hivemetastore) createDatasets(ctx context.Context, connector *dataextractor.Connector, workspaceName string) {
	metastore, err := m.getMetastoreConnection()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer metastore.Close()

	dbName := connector.GetWarehouseDatabaseName()

	for _, table := range connector.Tables {
		tableHiveName := table.GetHiveName()

		// check if dataset exists
		var res *http.Response
		res, err = m.httpClient.Get(fmt.Sprintf("%s/dataset?wid=%s&dataset=%s", m.config.dataspaceHost, dbName, tableHiveName), nil)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		var parsedResponse dataspaceApiResponse
		err = json.NewDecoder(res.Body).Decode(&parsedResponse)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Println(parsedResponse)

		if _, ok := parsedResponse.Data.(map[string]interface{}); ok {
			continue
		}
		// if there is no dataset then create
		hiveTable, err := metastore.Client.GetTable(ctx, dbName, tableHiveName)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		elements := make([]dataspaceElement, len(hiveTable.Sd.Cols))
		for i, col := range hiveTable.Sd.Cols {
			elements[i] = dataspaceElement{Name: col.Name, DataType: col.Type}
		}
		dataset := dataspaceDataset{
			WorkspaceId:   dbName,
			Name:          tableHiveName,
			Description:   connector.Name + " " + tableHiveName,
			Database:      dbName,
			Creator:       workspaceName,
			Frequency:     "daily",
			FrequencyCron: "0 0 0 * * ?",
			StartDate:     "",
			DatsetType:    "Structured Data",
			Elements:      elements,
		}
		if connector.Schedule != nil && connector.Schedule.CronSchedule != nil {
			dataset.FrequencyCron = connector.Schedule.CronSchedule.Expression
		}

		// create dataset
		reqBody, err := json.Marshal(dataset)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		header := http.Header{}
		header.Add("Content-Type", "application/json")
		res, err = m.httpClient.Post(fmt.Sprintf("%s/add-dataset", m.config.dataspaceHost), bytes.NewBuffer(reqBody), header)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		err = json.NewDecoder(res.Body).Decode(&parsedResponse)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Println(parsedResponse)
	}
}

func (m *hivemetastore) RefreshTablesPartitions(ctx context.Context, connector *dataextractor.Connector, workspaceName string) error {
	spark, err := m.getSparkConnection()
	if err != nil {
		return err
	}

	defer spark.Close()

	cursor := spark.Cursor()

	dbName := connector.GetWarehouseDatabaseName()

	cursor.Exec(ctx, fmt.Sprintf("USE %s", dbName))
	if cursor.Err != nil {
		return cursor.Err
	}

	for _, table := range connector.Tables {
		tableHiveName := table.GetHiveName()
		cursor.Exec(ctx, fmt.Sprintf("MSCK REPAIR TABLE `%s`", tableHiveName))
		if cursor.Err != nil {
			return cursor.Err
		}
	}

	cursor.Close()

	m.createDatasets(ctx, connector, workspaceName)

	return nil
}

type dataserviceApiResponse struct {
	ResponseCode string      `json:"reponseCode"`
	Success      bool        `json:"success"`
	Data         string      `json:"data"`
	Message      interface{} `json:"message"`
}

// TODO: this will be replaced by the dripper interface api
func (m *hivemetastore) createDataprofiles(ctx context.Context, connector *dataextractor.Connector) {
	dbName := connector.GetWarehouseDatabaseName()
	for _, table := range connector.Tables {
		tableHiveName := table.GetHiveName()
		url := fmt.Sprintf("%s/dataProfile?inputSchema=%s&inputTable=%s", m.config.dataserviceHost, dbName, tableHiveName)

		res, err := m.httpClient.Get(url, nil)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		var parsedResponse dataserviceApiResponse
		err = json.NewDecoder(res.Body).Decode(&parsedResponse)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Println(parsedResponse)
	}
}

func (m *hivemetastore) ProfileTables(ctx context.Context, connector *dataextractor.Connector) error {
	m.createDataprofiles(ctx, connector)

	metastore, err := m.getMetastoreConnection()
	if err != nil {
		return err
	}

	defer metastore.Close()

	spark, err := m.getSparkConnection()
	if err != nil {
		return err
	}

	defer spark.Close()

	cursor := spark.Cursor()

	dbName := connector.GetWarehouseDatabaseName()

	cursor.Exec(ctx, fmt.Sprintf("USE `%s`", dbName))
	if cursor.Err != nil {
		return cursor.Err
	}

	for _, table := range connector.Tables {
		tableHiveName := table.GetHiveName()
		hiveTable, err := metastore.Client.GetTable(ctx, dbName, tableHiveName)
		if err != nil {
			return err
		}

		for _, col := range hiveTable.Sd.Cols {
			query := fmt.Sprintf("ANALYZE TABLE `%s` COMPUTE STATISTICS FOR COLUMNS `%s`", tableHiveName, col.Name)
			fmt.Println(query)
			cursor.Exec(ctx, query)

			if col.Type == "bigint" || col.Type == "double" {
				var mean float64
				var median float64
				var variance float64
				var stddev float64
				var kurtosis float64

				query := fmt.Sprintf("SELECT MEAN(`%[1]v`), PERCENTILE(`%[1]v`, 0.5), VARIANCE(`%[1]v`), STDDEV(`%[1]v`), KURTOSIS(`%[1]v`) FROM `%[2]v`", col.Name, tableHiveName)
				fmt.Println(query)
				cursor.Exec(ctx, query)

				for cursor.HasMore(ctx) {
					cursor.FetchOne(ctx, &mean, &median, &variance, &stddev, &kurtosis)
					if cursor.Err != nil {
						return cursor.Err
					}
				}
				stats := map[string]interface{}{
					"mean":     mean,
					"median":   median,
					"variance": variance,
					"stddev":   stddev,
					"kurtosis": kurtosis,
				}

				statsJson, err := json.Marshal(stats)
				if err != nil {
					return err
				}

				// save number extra stats on column description
				col.Comment = string(statsJson)
				fmt.Println(col)
			}
		}

		err = metastore.Client.AlterTable(ctx, dbName, tableHiveName, hiveTable)
		if err != nil {
			return err
		}
	}

	cursor.Close()

	return nil
}
