package fixtures

import (
	"encoding/json"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/google/uuid"
	dataextractor "github.com/ubix/dripper/manager/repositories/data-extractor"
	airbyteclient "github.com/ubix/dripper/manager/repositories/data-extractor/airbyte/airbyte-client"
)

func GetBytesSynced() *int64 {
	n := rand.Int64()
	return &n
}

func GetConnectionScheduleUnits() int64 {
	return rand.Int64()
}

func getJobId() int64 {
	return rand.Int64()
}

func GetName() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 10)
	for i := range b {
		b[i] = letterBytes[rand.IntN(len(letterBytes))]
	}
	return string(b)
}

func GetNewUUID() uuid.UUID {
	return uuid.New()
}

func GetOperationsIds() []uuid.UUID {
	uuid1 := GetNewUUID()
	uuid2 := GetNewUUID()
	return []uuid.UUID{uuid1, uuid2}
}

func GetTime() time.Time {
	return time.Now()
}

func GetRecordsSynced() *int64 {
	n := rand.Int64()
	return &n
}

func GetTimeInSeconds() int64 {
	return time.Now().Unix()
}

var (
	airbyteStreamConfigurationSelected bool                   = true
	configType                         string                 = "sync"
	connectionScheduleBasicTimeUnit    string                 = "minutes"
	connectionScheduleCronExpression   string                 = "0 0 12 * * ?"
	connectionScheduleCronTimeZone     string                 = "America/Sao_Paulo"
	connectionStatus                   string                 = "active"
	cursorField                        []string               = []string{"sample_cursor_field"}
	destinationSyncMode                string                 = "append"
	namespace                          string                 = "test"
	jobStatus                          string                 = "completed"
	jsonSchema                         map[string]interface{} = map[string]interface{}{
		"properties": map[string]interface{}{
			"column1": map[string]interface{}{
				"type": "string",
			},
		},
		"type": "object",
	}
	namespaceDefinitionType           string                   = "customformat"
	namespaceFormat                   string                   = "${TEST_NAMESPACE}"
	prefix                            string                   = ""
	primaryKey                        [][]string               = [][]string{{"key1", "key2"}, {"key3", "key4"}}
	resourceRequirementsCpuLimit      string                   = "100"
	resourceRequirementsCpuRequest    string                   = "100"
	resourceRequirementsMemoryLimit   string                   = "100"
	resourceRequirementsMemoryRequest string                   = "100"
	sourceDefinedCursor               bool                     = true
	supportedSyncModes                []airbyteclient.SyncMode = []airbyteclient.SyncMode{"full_refresh", "incremental"}
	supportedSyncModesStr             []string                 = []string{"full_refresh", "incremental"}
	syncMode                          string                   = "incremental"
	DocumentationUrl                  string                   = "https://example.com/doc"
	IconUrl                           string                   = "https://example.com/icon.png"
	Configuration                     map[string]interface{}   = map[string]interface{}{"key": "value"}
	SourceStatus                      string                   = "active"
)

func GetDataExtractorConnectionReadWithBasicSchedule(
	connectionId uuid.UUID,
	destinationId uuid.UUID,
	name string,
	operationIds []uuid.UUID,
	sourceId uuid.UUID,
	connectionScheduleUnits int64) airbyteclient.ConnectionRead {
	return airbyteclient.ConnectionRead{
		ConnectionId:        connectionId,
		DestinationId:       destinationId,
		Name:                name,
		NamespaceDefinition: (*airbyteclient.NamespaceDefinitionType)(&namespaceDefinitionType),
		NamespaceFormat:     &namespaceFormat,
		OperationIds:        &operationIds,
		Prefix:              &prefix,
		ResourceRequirements: &airbyteclient.ResourceRequirements{
			CpuLimit:      &resourceRequirementsCpuLimit,
			CpuRequest:    &resourceRequirementsCpuRequest,
			MemoryLimit:   &resourceRequirementsMemoryLimit,
			MemoryRequest: &resourceRequirementsMemoryRequest,
		},
		ScheduleData: &airbyteclient.ConnectionScheduleData{
			BasicSchedule: &struct {
				TimeUnit airbyteclient.ConnectionScheduleDataBasicScheduleTimeUnit `json:"timeUnit"`
				Units    int64                                                     `json:"units"`
			}{
				TimeUnit: airbyteclient.ConnectionScheduleDataBasicScheduleTimeUnit(connectionScheduleBasicTimeUnit),
				Units:    int64(connectionScheduleUnits),
			},
		},
		SourceId: sourceId,
		Status:   airbyteclient.ConnectionStatus(connectionStatus),
		SyncCatalog: airbyteclient.AirbyteCatalog{
			Streams: []airbyteclient.AirbyteStreamAndConfiguration{
				{
					Config: airbyteclient.AirbyteStreamConfiguration{
						AliasName:           &name,
						CursorField:         &cursorField,
						DestinationSyncMode: airbyteclient.DestinationSyncMode(destinationSyncMode),
						PrimaryKey:          &primaryKey,
						Selected:            &airbyteStreamConfigurationSelected,
						SyncMode:            airbyteclient.SyncMode(syncMode),
					},
					Stream: airbyteclient.AirbyteStream{
						DefaultCursorField:      cursorField,
						JsonSchema:              jsonSchema,
						Name:                    name,
						Namespace:               &namespace,
						SourceDefinedCursor:     sourceDefinedCursor,
						SourceDefinedPrimaryKey: primaryKey,
						SupportedSyncModes:      supportedSyncModes,
					},
				},
			},
		},
	}
}

func GetDataExtractorConnectionReadWithCronSchedule(
	connectionId uuid.UUID,
	destinationId uuid.UUID,
	name string,
	operationIds []uuid.UUID,
	sourceId uuid.UUID,
	connectionScheduleUnits int64) airbyteclient.ConnectionRead {
	return airbyteclient.ConnectionRead{
		ConnectionId:        connectionId,
		DestinationId:       destinationId,
		Name:                name,
		NamespaceDefinition: (*airbyteclient.NamespaceDefinitionType)(&namespaceDefinitionType),
		NamespaceFormat:     &namespaceFormat,
		OperationIds:        &operationIds,
		Prefix:              &prefix,
		ResourceRequirements: &airbyteclient.ResourceRequirements{
			CpuLimit:      &resourceRequirementsCpuLimit,
			CpuRequest:    &resourceRequirementsCpuRequest,
			MemoryLimit:   &resourceRequirementsMemoryLimit,
			MemoryRequest: &resourceRequirementsMemoryRequest,
		},
		ScheduleData: &airbyteclient.ConnectionScheduleData{
			Cron: &struct {
				CronExpression string `json:"cronExpression"`
				CronTimeZone   string `json:"cronTimeZone"`
			}{
				CronExpression: connectionScheduleCronExpression,
				CronTimeZone:   connectionScheduleCronTimeZone,
			},
		},
		SourceId: sourceId,
		Status:   airbyteclient.ConnectionStatus(connectionStatus),
		SyncCatalog: airbyteclient.AirbyteCatalog{
			Streams: []airbyteclient.AirbyteStreamAndConfiguration{
				{
					Config: airbyteclient.AirbyteStreamConfiguration{
						AliasName:           &name,
						CursorField:         &cursorField,
						DestinationSyncMode: airbyteclient.DestinationSyncMode(destinationSyncMode),
						PrimaryKey:          &primaryKey,
						Selected:            &airbyteStreamConfigurationSelected,
						SyncMode:            airbyteclient.SyncMode(syncMode),
					},
					Stream: airbyteclient.AirbyteStream{
						DefaultCursorField:      cursorField,
						JsonSchema:              jsonSchema,
						Name:                    name,
						Namespace:               &namespace,
						SourceDefinedCursor:     sourceDefinedCursor,
						SourceDefinedPrimaryKey: primaryKey,
						SupportedSyncModes:      supportedSyncModes,
					},
				},
			},
		},
	}
}

func GetCreateConnectionWithBasicSchedule(
	connectionId uuid.UUID,
	destinationId uuid.UUID,
	name string,
	operationIds []uuid.UUID,
	sourceId uuid.UUID,
	connectionScheduleUnits int64) airbyteclient.WebBackendConnectionCreate {
	namespaceFormat := strings.ReplaceAll(sourceId.String(), "-", "")
	return airbyteclient.WebBackendConnectionCreate{
		SourceId:                     sourceId,
		DestinationId:                destinationId,
		Name:                         &name,
		NamespaceDefinition:          (*airbyteclient.NamespaceDefinitionType)(&namespaceDefinitionType),
		NamespaceFormat:              &namespaceFormat,
		NonBreakingChangesPreference: airbyteclient.PropagateColumns,
		ResourceRequirements: &airbyteclient.ResourceRequirements{
			CpuLimit:      &resourceRequirementsCpuLimit,
			CpuRequest:    &resourceRequirementsCpuRequest,
			MemoryLimit:   &resourceRequirementsMemoryLimit,
			MemoryRequest: &resourceRequirementsMemoryRequest,
		},
		Schedule: &airbyteclient.ConnectionSchedule{
			TimeUnit: airbyteclient.ConnectionScheduleTimeUnit(connectionScheduleBasicTimeUnit),
			Units:    int64(connectionScheduleUnits),
		},
		ScheduleData: &airbyteclient.ConnectionScheduleData{
			BasicSchedule: &struct {
				TimeUnit airbyteclient.ConnectionScheduleDataBasicScheduleTimeUnit `json:"timeUnit"`
				Units    int64                                                     `json:"units"`
			}{
				TimeUnit: airbyteclient.ConnectionScheduleDataBasicScheduleTimeUnit(connectionScheduleBasicTimeUnit),
				Units:    int64(connectionScheduleUnits),
			},
		},
		Status: airbyteclient.ConnectionStatus(connectionStatus),
		SyncCatalog: airbyteclient.AirbyteCatalog{
			Streams: []airbyteclient.AirbyteStreamAndConfiguration{
				{
					Config: airbyteclient.AirbyteStreamConfiguration{
						AliasName:           &name,
						CursorField:         &cursorField,
						DestinationSyncMode: airbyteclient.DestinationSyncMode(destinationSyncMode),
						PrimaryKey:          &primaryKey,
						Selected:            &airbyteStreamConfigurationSelected,
						SyncMode:            airbyteclient.SyncMode(syncMode),
					},
					Stream: airbyteclient.AirbyteStream{
						DefaultCursorField:      cursorField,
						JsonSchema:              jsonSchema,
						Name:                    name,
						Namespace:               &namespace,
						SourceDefinedCursor:     sourceDefinedCursor,
						SourceDefinedPrimaryKey: primaryKey,
						SupportedSyncModes:      supportedSyncModes,
					},
				},
			},
		},
	}

}

func GetConnectorBasicSchedule(
	Id uuid.UUID,
	destinationId uuid.UUID,
	name string,
	operationIds []uuid.UUID,
	sourceId uuid.UUID,
	connectionScheduleUnits int64) *dataextractor.Connector {
	connector := &dataextractor.Connector{
		ID:   &Id,
		Name: name,
		ResourceRequirements: &dataextractor.ResourceRequirements{
			CpuLimit:      &resourceRequirementsCpuLimit,
			CpuRequest:    &resourceRequirementsCpuRequest,
			MemoryLimit:   &resourceRequirementsMemoryLimit,
			MemoryRequest: &resourceRequirementsMemoryRequest,
		},
		Schedule: &dataextractor.Schedule{
			BasicSchedule: &dataextractor.BasicSchedule{
				TimeUnit: connectionScheduleBasicTimeUnit,
				Units:    connectionScheduleUnits,
			},
		},
		SourceId: &sourceId,
		Tables: []*dataextractor.Table{
			{
				DefaultCursorField:      cursorField,
				JsonSchema:              jsonSchema,
				Name:                    name,
				Namespace:               &namespace,
				SourceDefinedCursor:     sourceDefinedCursor,
				SourceDefinedPrimaryKey: primaryKey,
				SupportedSyncModes:      supportedSyncModesStr,
				SyncEnabled:             true,
				SelectedCursorField:     &cursorField,
			},
		},
		SourceConfiguration:      map[string]interface{}{},
		DestinationConfiguration: map[string]interface{}{},
	}
	connector.WarehouseDatabase = connector.GetWarehouseDatabaseName()
	return connector
}

func GetConnectorCronSchedule(
	Id uuid.UUID,
	destinationId uuid.UUID,
	name string,
	operationIds []uuid.UUID,
	sourceId uuid.UUID,
	connectionScheduleUnits int64) *dataextractor.Connector {
	connector := &dataextractor.Connector{
		ID:   &Id,
		Name: name,
		ResourceRequirements: &dataextractor.ResourceRequirements{
			CpuLimit:      &resourceRequirementsCpuLimit,
			CpuRequest:    &resourceRequirementsCpuRequest,
			MemoryLimit:   &resourceRequirementsMemoryLimit,
			MemoryRequest: &resourceRequirementsMemoryRequest,
		},
		Schedule: &dataextractor.Schedule{
			CronSchedule: &dataextractor.CronSchedule{
				Expression: connectionScheduleCronExpression,
				TimeZone:   connectionScheduleCronTimeZone,
			},
		},
		SourceId: &sourceId,
		Tables: []*dataextractor.Table{
			{
				DefaultCursorField:      cursorField,
				JsonSchema:              jsonSchema,
				Name:                    name,
				Namespace:               &namespace,
				SourceDefinedCursor:     sourceDefinedCursor,
				SourceDefinedPrimaryKey: primaryKey,
				SupportedSyncModes:      supportedSyncModesStr,
				SyncEnabled:             true,
				SelectedCursorField:     &cursorField,
			},
		},
		SourceConfiguration:      map[string]interface{}{},
		DestinationConfiguration: map[string]interface{}{},
	}
	connector.WarehouseDatabase = connector.GetWarehouseDatabaseName()
	return connector
}

func GetListConnectorsFixtures() (*airbyteclient.ConnectionReadList, []*dataextractor.Connector) {
	connectionId := GetNewUUID()
	destinationId := GetNewUUID()
	name := GetName()
	operationsIds := GetOperationsIds()
	sourceId := GetNewUUID()
	connectionScheduleUnits := GetConnectionScheduleUnits()
	expectedConnectors := make([]*dataextractor.Connector, 2)
	expectedConnectors[0] = GetConnectorBasicSchedule(
		connectionId,
		destinationId,
		name,
		operationsIds,
		sourceId,
		connectionScheduleUnits,
	)
	expectedConnectors[1] = GetConnectorCronSchedule(
		connectionId,
		destinationId,
		name,
		operationsIds,
		sourceId,
		connectionScheduleUnits,
	)

	return &airbyteclient.ConnectionReadList{
			Connections: []airbyteclient.ConnectionRead{
				GetDataExtractorConnectionReadWithBasicSchedule(
					connectionId,
					destinationId,
					name,
					operationsIds,
					sourceId,
					connectionScheduleUnits,
				),
				GetDataExtractorConnectionReadWithCronSchedule(
					connectionId,
					destinationId,
					name,
					operationsIds,
					sourceId,
					connectionScheduleUnits,
				),
			},
		},
		expectedConnectors
}

func GetListConnectorsWithoutSchedule() (*airbyteclient.ConnectionReadList, []*dataextractor.Connector) {
	connectionId := GetNewUUID()
	destinationId := GetNewUUID()
	name := GetName()
	operationsIds := GetOperationsIds()
	sourceId := GetNewUUID()
	connectionScheduleUnits := GetConnectionScheduleUnits()

	de := &airbyteclient.ConnectionReadList{
		Connections: []airbyteclient.ConnectionRead{
			GetDataExtractorConnectionReadWithBasicSchedule(
				connectionId,
				destinationId,
				name,
				operationsIds,
				sourceId,
				connectionScheduleUnits,
			),
		},
	}

	de.Connections[0].ScheduleData = nil

	expectedConnector := GetConnectorBasicSchedule(
		connectionId,
		destinationId,
		name,
		operationsIds,
		sourceId,
		connectionScheduleUnits,
	)
	expectedConnector.Schedule = &dataextractor.Schedule{
		BasicSchedule: nil,
		CronSchedule:  nil,
	}

	expectedConnectorsList := make([]*dataextractor.Connector, 1)
	expectedConnectorsList[0] = expectedConnector

	return de, expectedConnectorsList
}

const (
	inputJsonSchema = `{
    "type": "object",
    "properties": {
      "col_obj": {
        "type": "object",
        "properties": {
          "col_1": {
            "type": ["null", "string"]
          },
          "col_2": {
            "type": "string",
            "format": "date"
          },
          "col_3": {
            "airbyte_type": "integer",
            "type": "number"
          },
          "col_without_type": {},
          "col_empty_obj": {
            "type": "object"
          },
          "col_array_without_items": {
            "type": "array"
          }
        }
      },
      "column_arr": {
        "type": "array",
        "items": {
          "type": "str"
        }
      },
      "column_anyof": {
        "anyOf": [
          {
            "type": "object",
            "properties": {
              "col_ok_1": {
                "type": "str"
              },
              "col_not_ok": {
                "body_html": "str"
              }
            }
          },
          {
            "type": "object",
            "properties": {
              "col_ok_2": {
                "type": "str"
              },
              "col_not_ok": {
                "body_html": "str"
              }
            }
          }
        ]
      }
    }
  }`
	expectedJsonSchema = `{
    "type": "object",
    "properties": {
      "col_obj": {
        "type": "object",
        "properties": {
          "col_1": {
            "type": "string"
          },
          "col_2": {
            "type": "string",
            "format": "date"
          },
          "col_3": {
            "airbyte_type": "integer",
            "type": "number"
          }
        }
      },
      "column_arr": {
        "type": "array",
        "items": {
          "type": "str"
        }
      },
      "column_anyof": {
        "anyOf": [
          {
            "type": "object",
            "properties": {
              "col_ok_1": {
                "type": "str"
              }
            }
          },
          {
            "type": "object",
            "properties": {
              "col_ok_2": {
                "type": "str"
              }
            }
          }
        ]
      }
    }
  }`
)

func GetJsonSchemaFixtures() (map[string]interface{}, map[string]interface{}) {
	input := map[string]interface{}{}
	expected := map[string]interface{}{}

	json.Unmarshal([]byte(inputJsonSchema), &input)
	json.Unmarshal([]byte(expectedJsonSchema), &expected)

	return input, expected
}
