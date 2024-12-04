// +build all e2e

package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/beltran/gohive"
	"github.com/beltran/gohive/hive_metastore"

	airbyteclient "github.com/ubix/dripper/manager/repositories/data-extractor/airbyte/airbyte-client"
	"github.com/ubix/dripper/manager/test/e2e/configs"
	restapiclient "github.com/ubix/dripper/manager/test/e2e/rest-api-client"
	schemavalidator "github.com/ubix/dripper/manager/test/e2e/schema-validator"
)

type testClient struct {
	client         *restapiclient.ClientWithResponses
	airbyte        *airbyteclient.ClientWithResponses
	metastore      *gohive.HiveMetastoreClient
	e2eSourceDefId uuid.UUID
	s3DestDefId    uuid.UUID
	organizationId uuid.UUID
}

// we rewrite the api error to avoid using the one in production and bias our test
type apiError struct {
	statusCode int
}

func (e *apiError) Error() string {
	return fmt.Sprintf("API failed with status: %d", e.statusCode)
}

func newApiError(statusCode int) *apiError {
	return &apiError{statusCode}
}

func buildClient(config *configs.TestConfig) (*testClient, error) {
	host := config.RestAPIHost
	client, err := restapiclient.NewClientWithResponses(host)
	if err != nil {
		return nil, err
	}

	airbyte, err := airbyteclient.NewClientWithResponses(config.AirbyteAPIHost)
	if err != nil {
		return nil, err
	}

	e2eSourceDefId, err := uuid.Parse(config.E2ESourceDefID)
	if err != nil {
		return nil, err
	}

	s3DestDefId, err := uuid.Parse(config.S3DestDefID)
	if err != nil {
		return nil, err
	}

	organizationId, err := uuid.Parse(config.DefaultOrganizationID)
	if err != nil {
		return nil, err
	}

	metastoreConfig := gohive.NewMetastoreConnectConfiguration()
	metastore, err := gohive.ConnectToMetastore(config.MetastoreHost, config.MetastorePort, "NOSASL", metastoreConfig)
	if err != nil {
		return nil, err
	}

	return &testClient{
		client:         client,
		airbyte:        airbyte,
		metastore:      metastore,
		e2eSourceDefId: e2eSourceDefId,
		s3DestDefId:    s3DestDefId,
		organizationId: organizationId,
	}, nil
}

func (c *testClient) cleaunpWorkspaces(ctx context.Context) error {
	listWorkspacesResponse, err := c.airbyte.ListWorkspacesWithResponse(ctx)
	if err != nil {
		return err
	}
	if listWorkspacesResponse.StatusCode() != http.StatusOK {
		return newApiError(listWorkspacesResponse.StatusCode())
	}

	for _, workspace := range listWorkspacesResponse.JSON200.Workspaces {
		res, err := c.airbyte.DeleteWorkspace(
			ctx,
			airbyteclient.WorkspaceIdRequestBody{WorkspaceId: workspace.WorkspaceId},
		)
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
			return newApiError(res.StatusCode)
		}
	}

	return nil
}

func (c *testClient) setupWebhookNotifications(ctx context.Context, workspaceId uuid.UUID) error {
	notifications := []airbyteclient.NotificationType{"slack"}
	workspaceUpdate := airbyteclient.WorkspaceUpdate{
		WorkspaceId: workspaceId,
		NotificationSettings: &airbyteclient.NotificationSettings{
			SendOnSuccess: airbyteclient.NotificationSettingsSendOn{
				NotificationType: &notifications,
				SlackConfiguration: &airbyteclient.SlackNotificationConfiguration{
					Webhook: "http://host.docker.internal:8080/airbyte_internal/sync_webhook",
				},
			},
		},
	}

	res, err := c.airbyte.UpdateWorkspaceWithResponse(ctx, workspaceUpdate)
	if err != nil {
		return err
	}

	if res.StatusCode() != http.StatusOK {
		return newApiError(res.StatusCode())
	}

	return nil
}

func (c *testClient) cleanupAndSetupWorkspace(ctx context.Context) (*uuid.UUID, error) {
	err := c.cleaunpWorkspaces(ctx)
	if err != nil {
		return nil, err
	}

	// create new workspace to ensure it is the only and thus the default
	createWorkspaceResponse, err := c.airbyte.CreateWorkspaceWithResponse(
		ctx,
		airbyteclient.WorkspaceCreate{
			Name:           "e2e_test",
			OrganizationId: c.organizationId,
		},
	)
	if err != nil {
		return nil, err
	}
	// the api may either return 200 our 204 upon workspace creation ok
	if createWorkspaceResponse.StatusCode() != http.StatusOK && createWorkspaceResponse.StatusCode() != http.StatusNoContent {
		return nil, newApiError(createWorkspaceResponse.StatusCode())
	}

	dbs, err := c.metastore.Client.GetAllDatabases(ctx)
	if err != nil {
		return nil, err
	}

	for _, db := range dbs {
		if db == "default" {
			continue
		}
		err = c.metastore.Client.DropDatabase(ctx, db, true, true)
		if err != nil {
			return nil, err
		}
	}

	workspaceId := createWorkspaceResponse.JSON200.WorkspaceId

	err = c.setupWebhookNotifications(ctx, workspaceId)
	if err != nil {
		return nil, err
	}

	return &workspaceId, nil
}

func (c *testClient) createE2ESource(ctx context.Context, name string, workspaceId uuid.UUID) (*uuid.UUID, error) {
	res, err := c.airbyte.CreateSourceWithResponse(ctx, airbyteclient.SourceCreate{
		ConnectionConfiguration: map[string]interface{}{
			"max_records": 100,
			"type":        "INFINITE_FEED",
		},
		WorkspaceId:        workspaceId,
		Name:               name,
		SourceDefinitionId: c.e2eSourceDefId,
	})
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, newApiError(res.StatusCode())
	}

	return &res.JSON200.SourceId, nil
}

func (c *testClient) createOrUpdateS3Dest(ctx context.Context, config *configs.TestConfig, name string, workspaceId uuid.UUID) (*uuid.UUID, error) {
	listRes, err := c.airbyte.ListDestinationsForWorkspaceWithResponse(
		ctx,
		airbyteclient.WorkspaceIdRequestBody{WorkspaceId: workspaceId},
	)
	if err != nil {
		return nil, err
	}

	if listRes.StatusCode() != http.StatusOK {
		return nil, newApiError(listRes.StatusCode())
	}

	var destinationId *uuid.UUID

	if len(listRes.JSON200.Destinations) == 1 {
		destination := listRes.JSON200.Destinations[0]
		res, err := c.airbyte.UpdateDestinationWithResponse(ctx, airbyteclient.DestinationUpdate{
			ConnectionConfiguration: config.S3MinioConfig,
			Name:                    destination.Name,
			DestinationId:           destination.DestinationId,
		})
		if err != nil {
			return nil, err
		}
		if res.StatusCode() != http.StatusOK {
			return nil, newApiError(res.StatusCode())
		}
		destinationId = &destination.DestinationId
	} else {
		res, err := c.airbyte.CreateDestinationWithResponse(ctx, airbyteclient.CreateDestinationJSONRequestBody{
			ConnectionConfiguration: config.S3MinioConfig,
			WorkspaceId:             workspaceId,
			Name:                    name,
			DestinationDefinitionId: c.s3DestDefId,
		})
		if err != nil {
			return nil, err
		}
		if res.StatusCode() != http.StatusOK {
			return nil, newApiError(res.StatusCode())
		}

		destinationId = &res.JSON200.DestinationId
	}

	return destinationId, nil
}

const (
	rawJsonSchema = `{
    "type": "object",
    "properties": {
      "col_str": {
        "type": ["null", "string"]
      },
      "col_str_date": {
        "type": "string",
        "format": "date"
      },
      "col_str_datetime": {
        "type": "string",
        "format": "date-time"
      },
      "col_int": {
        "type": "integer"
      },
      "col_dec": {
        "type": "number"
      }
    }
  }`
)

func getTable() interface{} {
	jsonSchema := map[string]interface{}{}
	json.Unmarshal([]byte(rawJsonSchema), &jsonSchema)
	return map[string]interface{}{
		"name":                    "test",
		"jsonSchema":              jsonSchema,
		"supportedSyncModes":      []string{"full_refresh"},
		"sourceDefinedCursor":     false,
		"defaultCursorField":      []string{},
		"sourceDefinedPrimaryKey": []string{},
		"syncEnabled":             true,
	}
}

func getAirbyteSyncCatalog() airbyteclient.AirbyteCatalog {
	config := map[string]interface{}{
		"syncMode":              "full_refresh",
		"cursorField":           []string{},
		"destinationSyncMode":   "overwrite",
		"primaryKey":            []string{},
		"aliasName":             "test",
		"selected":              true,
		"suggested":             false,
		"fieldSelectionEnabled": false,
		"selectedFields":        []string{},
	}
	stream := getTable()
	streamAndConfig := map[string]interface{}{
		"stream": stream,
		"config": config,
	}

	var streams []interface{}
	streams = append(streams, streamAndConfig)

	syncCatalog := map[string]interface{}{"streams": streams}

	jsonBytes, _ := json.Marshal(syncCatalog)
	airbyteSyncCatalog := airbyteclient.AirbyteCatalog{}
	json.Unmarshal(jsonBytes, &airbyteSyncCatalog)

	return airbyteSyncCatalog
}

func getTableAsStruct() restapiclient.Table {
	jsonBytes, _ := json.Marshal(getTable())
	table := restapiclient.Table{}
	json.Unmarshal(jsonBytes, &table)
	return table
}

func (c *testClient) createConnection(ctx context.Context, name string, sourceId uuid.UUID, destinationId uuid.UUID) (*uuid.UUID, error) {
	namespaceDefinition := airbyteclient.Customformat
	namespaceFormat := sourceId.String()
	syncCatalog := getAirbyteSyncCatalog()
	res, err := c.airbyte.WebBackendCreateConnectionWithResponse(ctx, airbyteclient.WebBackendConnectionCreate{
		SourceId:                     sourceId,
		DestinationId:                destinationId,
		ScheduleType:                 airbyteclient.ConnectionScheduleTypeManual,
		Status:                       "active",
		Name:                         &name,
		SyncCatalog:                  syncCatalog,
		NamespaceDefinition:          &namespaceDefinition,
		NamespaceFormat:              &namespaceFormat,
		NonBreakingChangesPreference: airbyteclient.PropagateColumns,
	})
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, newApiError(res.StatusCode())
	}

	return &res.JSON200.ConnectionId, nil
}

func (c *testClient) triggerJob(ctx context.Context, connectionId uuid.UUID) (*int64, error) {
	res, err := c.airbyte.PublicCreateJobWithResponse(ctx, airbyteclient.JobCreateRequest{
		ConnectionId: connectionId,
		JobType:      airbyteclient.JobTypeEnumSync,
	})
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, newApiError(res.StatusCode())
	}

	return &res.JSON200.JobId, nil
}

const (
	singleTenantBucket = "warehouse"
)

func TestE2E(t *testing.T) {
	var config *configs.TestConfig
	var err error

	config, err = configs.LoadTestConfig()
	if err != nil {
		panic(err)
	}

	c, err := buildClient(config)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	ctx := context.Background()
	workspaceId := &uuid.Nil
	destinationId := &uuid.Nil

	t.Run("CreateWorkspace", func(t *testing.T) {
		err = c.cleaunpWorkspaces(ctx)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		res, err := c.client.CreateWorkspace(
			ctx,
			restapiclient.CreateWorkspaceRequestBody{
				Tenant: singleTenantBucket,
				Bucket: singleTenantBucket,
			},
		)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.NotNil(t, res)
		assert.Equal(t, http.StatusCreated, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		bodyReader := bytes.NewBuffer(body)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "Workspace", io.NopCloser(bodyReader))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)

		// Validate the ticketId endpoint (status:: pending)
		var workspace restapiclient.Workspace
		err = json.Unmarshal(body, &workspace)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		workspaceId = &workspace.Id
	})

	if workspaceId == &uuid.Nil {
		workspaceId, err = c.cleanupAndSetupWorkspace(ctx)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		err = c.setupWebhookNotifications(ctx, *workspaceId)
		if err != nil {
			assert.Fail(t, err.Error())
		}
	}

	destinationId, err = c.createOrUpdateS3Dest(ctx, config, "default_dest", *workspaceId)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	t.Run("GetWorkspace", func(t *testing.T) {
		res, err := c.client.GetWorkspace(ctx, *workspaceId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.NotNil(t, res)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "Workspace", res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)
	})

	t.Run("ListWorkspaces", func(t *testing.T) {
		res, err := c.client.ListWorkspaces(ctx, nil)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.NotNil(t, res)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "WorkspaceList", res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)
	})

	t.Run("ListSourcesDefinitions", func(t *testing.T) {
		res, err := c.client.ListSourcesDefinitions(context.Background(), *workspaceId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.NotNil(t, res)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "SourceDefinitionList", res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)
	})

	t.Run("GetSourceDefinitionConfiguration", func(t *testing.T) {
		listSourcesDefsRes, err := c.client.ListSourcesDefinitionsWithResponse(context.Background(), *workspaceId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		sourceDefId := (*listSourcesDefsRes.JSON200)[0].SourceDefinitionId

		schemaRes, err := c.client.GetSourceDefinitionConfiguration(ctx, *workspaceId, sourceDefId)

		assert.Equal(t, http.StatusOK, schemaRes.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "SourceDefinitionConfiguration", schemaRes.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)
	})

	t.Run("CreateSourceFromDefinitionError", func(t *testing.T) {
		sourceName := "create_source_e2e_test_error"
		sourceConf := map[string]interface{}{
			"max_messages":        100,
			"type":                "CONTINUOUS_FEED",
			"seed":                0,
			"message_interval_ms": 0,
			"mock_catalog": map[string]interface{}{
				"stream_duplication": 1,
				"stream_name":        "test",
				"stream_schema":      "bad_schema",
				"type":               "SINGLE_STREAM",
			},
		}
		res, err := c.client.CreateSourceFromDefinition(ctx, *workspaceId, c.e2eSourceDefId, restapiclient.CreateSourceRequestBody{
			Name:          &sourceName,
			Configuration: &sourceConf,
		})

		if err != nil {
			assert.Fail(t, err.Error())
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		bodyReader := bytes.NewBuffer(body)

		assert.Equal(t, http.StatusAccepted, res.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "AsyncOperationResponse", io.NopCloser(bodyReader))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)

		// Validate the ticketId endpoint (status:: pending)
		var asyncResponse restapiclient.AsyncOperationResponse
		err = json.Unmarshal(body, &asyncResponse)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		poolingRes, err := c.client.GetCreateSourceFromDefinitionStatus(ctx, *workspaceId, c.e2eSourceDefId, asyncResponse.TicketId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		body, err = io.ReadAll(poolingRes.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		bodyReader = bytes.NewBuffer(body)

		assert.Equal(t, http.StatusOK, poolingRes.StatusCode)

		isValid, err = schemavalidator.Validate(config.OpenAPISchemaPath, "CreateSourceStatus", io.NopCloser(bodyReader))
		assert.True(t, isValid)

		// Validate the ticketId endpoint (status:: finished)
		var creationStatus restapiclient.CreateSourceStatus
		err = json.Unmarshal(body, &creationStatus)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		status := creationStatus.Status
		for status == restapiclient.AsyncRequestStatusEnumPending {
			time.Sleep(1 * time.Second)
			poolingRes, err = c.client.GetCreateSourceFromDefinitionStatus(ctx, *workspaceId, c.e2eSourceDefId, asyncResponse.TicketId)
			if err != nil {
				assert.Fail(t, err.Error())
			}

			body, err = io.ReadAll(poolingRes.Body)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			err = json.Unmarshal(body, &creationStatus)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			status = creationStatus.Status
		}

		assert.Equal(t, restapiclient.AsyncRequestStatusEnumFailed, status)
		assert.Nil(t, creationStatus.SourceId)
		assert.True(t, *creationStatus.Error != "")
	})

	var sourceId *uuid.UUID

	t.Run("CreateSourceFromDefinitionOk", func(t *testing.T) {
		// Validate the async response from the CreateSourceFromDefinition
		sourceName := "create_source_e2e_test"
		sourceConf := map[string]interface{}{
			"max_messages":        100,
			"type":                "CONTINUOUS_FEED",
			"seed":                0,
			"message_interval_ms": 0,
			"mock_catalog": map[string]interface{}{
				"stream_duplication": 1,
				"stream_name":        "test",
				"stream_schema":      rawJsonSchema,
				"type":               "SINGLE_STREAM",
			},
		}
		res, err := c.client.CreateSourceFromDefinition(ctx, *workspaceId, c.e2eSourceDefId, restapiclient.CreateSourceRequestBody{
			Name:          &sourceName,
			Configuration: &sourceConf,
		})

		if err != nil {
			assert.Fail(t, err.Error())
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		bodyReader := bytes.NewBuffer(body)

		assert.Equal(t, http.StatusAccepted, res.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "AsyncOperationResponse", io.NopCloser(bodyReader))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)

		// Validate the ticketId endpoint (status:: pending)
		var asyncResponse restapiclient.AsyncOperationResponse
		err = json.Unmarshal(body, &asyncResponse)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		poolingRes, err := c.client.GetCreateSourceFromDefinitionStatus(ctx, *workspaceId, c.e2eSourceDefId, asyncResponse.TicketId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		body, err = io.ReadAll(poolingRes.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		bodyReader = bytes.NewBuffer(body)

		assert.Equal(t, http.StatusOK, poolingRes.StatusCode)

		isValid, err = schemavalidator.Validate(config.OpenAPISchemaPath, "CreateSourceStatus", io.NopCloser(bodyReader))
		assert.True(t, isValid)

		// Validate the ticketId endpoint (status:: finished)
		var creationStatus restapiclient.CreateSourceStatus
		err = json.Unmarshal(body, &creationStatus)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		status := creationStatus.Status
		for status == restapiclient.AsyncRequestStatusEnumPending {
			time.Sleep(1 * time.Second)
			poolingRes, err = c.client.GetCreateSourceFromDefinitionStatus(ctx, *workspaceId, c.e2eSourceDefId, asyncResponse.TicketId)
			if err != nil {
				assert.Fail(t, err.Error())
			}

			body, err = io.ReadAll(poolingRes.Body)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			err = json.Unmarshal(body, &creationStatus)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			status = creationStatus.Status
		}
		sourceId = creationStatus.SourceId

		assert.Equal(t, restapiclient.AsyncRequestStatusEnumFinished, status)
		assert.NotNil(t, sourceId)
	})

	if sourceId == nil {
		sourceId, err = c.createE2ESource(ctx, "test_source", *workspaceId)
		if err != nil {
			assert.Fail(t, err.Error())
		}
	}

	t.Run("GetSourceTables", func(t *testing.T) {
		res, err := c.client.GetSourceTables(ctx, *workspaceId, *sourceId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		bodyReader := bytes.NewBuffer(body)

		assert.Equal(t, http.StatusOK, res.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "AsyncOperationResponse", io.NopCloser(bodyReader))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)

		// Validate the ticketId endpoint (status:: pending)
		var asyncResponse restapiclient.AsyncOperationResponse
		err = json.Unmarshal(body, &asyncResponse)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		poolingRes, err := c.client.GetGetSourceTablesStatus(ctx, *workspaceId, *sourceId, asyncResponse.TicketId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		body, err = io.ReadAll(poolingRes.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		bodyReader = bytes.NewBuffer(body)

		assert.Equal(t, http.StatusOK, poolingRes.StatusCode)

		isValid, err = schemavalidator.Validate(config.OpenAPISchemaPath, "GetSourceTablesStatus", io.NopCloser(bodyReader))
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.True(t, isValid)

		// Validate the ticketId endpoint (status:: finished)
		var getSourceTablesStatus restapiclient.GetSourceTablesStatus
		err = json.Unmarshal(body, &getSourceTablesStatus)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		status := getSourceTablesStatus.Status
		for status == restapiclient.GetSourceTablesStatusStatusPending {
			time.Sleep(1 * time.Second)
			poolingRes, err = c.client.GetGetSourceTablesStatus(ctx, *workspaceId, *sourceId, asyncResponse.TicketId)
			if err != nil {
				assert.Fail(t, err.Error())
			}

			body, err = io.ReadAll(poolingRes.Body)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			err = json.Unmarshal(body, &getSourceTablesStatus)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			status = getSourceTablesStatus.Status
		}
		tables := getSourceTablesStatus.Tables

		assert.Equal(t, restapiclient.GetSourceTablesStatusStatusFinished, status)
		assert.NotNil(t, tables)
	})

	t.Run("ListSources", func(t *testing.T) {
		res, err := c.client.ListSources(ctx, *workspaceId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.NotNil(t, res)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "SourceList", res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)
	})

	t.Run("GetSource", func(t *testing.T) {
		res, err := c.client.GetSource(ctx, *workspaceId, *sourceId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.NotNil(t, res)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "Source", res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)
	})

	t.Run("UpdateSource", func(t *testing.T) {
		// Validate the async response from the CreateSourceFromDefinition
		sourceToUpdate, err := c.createE2ESource(ctx, "update_source_e2e_test", *workspaceId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		sourceConf := map[string]interface{}{
			"max_messages":        100,
			"type":                "CONTINUOUS_FEED",
			"seed":                0,
			"message_interval_ms": 0,
			"mock_catalog": map[string]interface{}{
				"stream_duplication": 1,
				"stream_name":        "test",
				"stream_schema":      rawJsonSchema,
				"type":               "SINGLE_STREAM",
			},
		}

		updatedName := "update_source_e2e_test_updated"
		res, err := c.client.UpdateSource(ctx, *workspaceId, *sourceToUpdate, restapiclient.CreateSourceRequestBody{
			Name:          &updatedName,
			Configuration: &sourceConf,
		})
		if err != nil {
			assert.Fail(t, err.Error())
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		bodyReader := bytes.NewBuffer(body)

		assert.Equal(t, http.StatusAccepted, res.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "AsyncOperationResponse", io.NopCloser(bodyReader))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)

		// Validate the ticketId endpoint (status:: pending)
		var asyncResponse restapiclient.AsyncOperationResponse
		err = json.Unmarshal(body, &asyncResponse)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		poolingRes, err := c.client.GetUpdateSourceStatus(ctx, *workspaceId, *sourceToUpdate, asyncResponse.TicketId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		body, err = io.ReadAll(poolingRes.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		bodyReader = bytes.NewBuffer(body)

		assert.Equal(t, http.StatusOK, poolingRes.StatusCode)

		isValid, err = schemavalidator.Validate(config.OpenAPISchemaPath, "UpdateSourceStatus", io.NopCloser(bodyReader))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)

		// Validate the ticketId endpoint (status:: finished)
		var updateStatus restapiclient.UpdateSourceStatus
		err = json.Unmarshal(body, &updateStatus)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		status := updateStatus.Status
		for status == restapiclient.AsyncRequestStatusEnumPending {
			time.Sleep(1 * time.Second)
			poolingRes, err = c.client.GetUpdateSourceStatus(ctx, *workspaceId, *sourceToUpdate, asyncResponse.TicketId)
			if err != nil {
				assert.Fail(t, err.Error())
			}

			body, err = io.ReadAll(poolingRes.Body)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			err = json.Unmarshal(body, &updateStatus)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			status = updateStatus.Status
		}
		updatedSource := updateStatus.Source

		assert.Equal(t, restapiclient.AsyncRequestStatusEnumFinished, status)
		assert.NotNil(t, updatedSource)
	})

	t.Run("DeleteSource", func(t *testing.T) {
		sourceIdToDelete, err := c.createE2ESource(ctx, "delete_source_source", *workspaceId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		res, err := c.client.DeleteSource(context.Background(), *workspaceId, *sourceIdToDelete)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.NotNil(t, res)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	var connectorId *uuid.UUID

	t.Run("CreateConnectorFromSource", func(t *testing.T) {
		connectorName := "create_connector_from_source_test"

		table := getTableAsStruct()

		res, err := c.client.CreateConnectorFromSource(ctx, *workspaceId, *sourceId, restapiclient.CreateConnectorFromSourceJSONRequestBody{
			Name:         connectorName,
			ScheduleType: restapiclient.Manual,
			Tables:       []restapiclient.Table{table},
		})
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.Equal(t, http.StatusCreated, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		bodyReader := bytes.NewBuffer(body)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "Connector", io.NopCloser(bodyReader))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)

		connector := restapiclient.Connector{}
		err = json.Unmarshal(body, &connector)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		connectorId = &connector.Id

		database, err := c.metastore.Client.GetDatabase(ctx, connector.WarehouseDatabase)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		tables, err := c.metastore.Client.GetAllTables(ctx, database.Name)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.True(t, len(tables) == 1)

		cols, err := c.metastore.Client.GetFields(ctx, database.Name, tables[0])
		if err != nil {
			assert.Fail(t, err.Error())
		}

		// sort the output to avoid random json keys parser error
		sort.Slice(cols, func(i, j int) bool {
			return cols[i].Name < cols[j].Name
		})

		expectedCols := []*hive_metastore.FieldSchema{
			&hive_metastore.FieldSchema{
				Name: "col_dec",
				Type: "double",
			},
			&hive_metastore.FieldSchema{
				Name: "col_int",
				Type: "bigint",
			},
			&hive_metastore.FieldSchema{
				Name: "col_str",
				Type: "string",
			},
			&hive_metastore.FieldSchema{
				Name: "col_str_date",
				Type: "date",
			},
			&hive_metastore.FieldSchema{
				Name: "col_str_datetime",
				Type: "timestamp",
			},
		}

		assert.Equal(t, len(expectedCols), len(cols))
		for i := range expectedCols {
			assert.Equal(t, *expectedCols[i], *cols[i])
		}
	})

	if connectorId == nil {
		connectorId, err = c.createConnection(ctx, "get_connectors_connection", *sourceId, *destinationId)
		if err != nil {
			assert.Fail(t, err.Error())
		}
	}

	t.Run("UpdateConnector", func(t *testing.T) {
		connectorName := "update_connector"

		table := getTableAsStruct()

		res, err := c.client.UpdateConnector(ctx, *workspaceId, *connectorId, restapiclient.CreateConnectorFromSourceJSONRequestBody{
			Name:         connectorName,
			ScheduleType: restapiclient.Manual,
			Tables:       []restapiclient.Table{table},
		})
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		bodyReader := bytes.NewBuffer(body)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "Connector", io.NopCloser(bodyReader))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)

		connector := restapiclient.Connector{}
		err = json.Unmarshal(body, &connector)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.Equal(t, connectorName, connector.Name)
	})

	t.Run("GetConnector", func(t *testing.T) {
		res, err := c.client.GetConnector(context.Background(), *workspaceId, *connectorId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.NotNil(t, res)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "Connector", res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)
	})

	t.Run("ListConnectors", func(t *testing.T) {
		limit := 100
		offset := 0
		res, err := c.client.ListConnectors(context.Background(), *workspaceId, &restapiclient.ListConnectorsParams{Limit: &limit, Offset: &offset})
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.NotNil(t, res)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "ConnectorsPaginated", res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)
	})

	t.Run("DeleteConnector", func(t *testing.T) {
		connId, err := c.createConnection(ctx, "delete_connectors_connection", *sourceId, *destinationId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		res, err := c.client.DeleteConnector(context.Background(), *workspaceId, *connId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.NotNil(t, res)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	var jobId *int64

	t.Run("TriggerConnectorJob", func(t *testing.T) {
		res, err := c.client.TriggerJobRun(context.Background(), *workspaceId, *connectorId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		bodyReader := bytes.NewBuffer(body)

		assert.NotNil(t, res)
		assert.Equal(t, http.StatusCreated, res.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "Job", io.NopCloser(bodyReader))
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)

		job := restapiclient.Job{}
		err = json.Unmarshal(body, &job)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		jobId = job.Id
	})

	if jobId == nil {
		jobId, err = c.triggerJob(ctx, *connectorId)
		if err != nil {
			assert.Fail(t, err.Error())
		}
	}

	t.Run("ListConnectorJobs", func(t *testing.T) {
		limit := 100
		offset := 0
		res, err := c.client.ListJobs(context.Background(), *workspaceId, *connectorId, &restapiclient.ListJobsParams{Limit: &limit, Offset: &offset})
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.NotNil(t, res)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "JobList", res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)
	})

	t.Run("GetConnectorJobDetails", func(t *testing.T) {
		res, err := c.client.GetJob(context.Background(), *workspaceId, *connectorId, *jobId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.NotNil(t, res)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "Job", res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)
	})

	t.Run("GetConnectorJobLogs", func(t *testing.T) {
		limit := 100
		offset := 0
		filters := &restapiclient.GetJobLogsParams{Limit: &limit, Offset: &offset}
		res, err := c.client.GetJobLogs(context.Background(), *workspaceId, *connectorId, *jobId, filters)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.NotNil(t, res)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		isValid, err := schemavalidator.Validate(config.OpenAPISchemaPath, "JobLog", res.Body)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.True(t, isValid)
	})

	t.Run("TriggerProfilling", func(t *testing.T) {
		// TODO: await for job success in x time and check profilling
	})

	t.Run("DeleteConnector", func(t *testing.T) {
		connId, err := c.createConnection(ctx, "delete_connectors_connection", *sourceId, *destinationId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		res, err := c.client.DeleteConnector(context.Background(), *workspaceId, *connId)
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.NotNil(t, res)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		// TODO: add metastore check here
	})

	// We don't delete the workspace after since this would break the web ui for local tests
	//  res, err := c.airbyte.DeleteWorkspace(
	// 	ctx,
	// 	airbyteclient.WorkspaceIdRequestBody{WorkspaceId: workspaceId},
	// )
	// if err != nil {
	// 	assert.Fail(t, err.Error())
	// }
	// if res.StatusCode != http.Status {
	// 	assert.Fail(t, airbyteApiError(res.StatusCode))
	// }
}
