package airbyte

import (
	"github.com/google/uuid"

	dataextractor "github.com/ubix/dripper/manager/repositories/data-extractor"
	airbyteclient "github.com/ubix/dripper/manager/repositories/data-extractor/airbyte/airbyte-client"
)

func ParseSourceReadList(sourceReadList *airbyteclient.SourceReadList) []*dataextractor.Source {
	sourceList := make([]*dataextractor.Source, len(sourceReadList.Sources))
	for i, sourceRead := range sourceReadList.Sources {
		sourceList[i] = ParseSourceRead(sourceRead)
		sourceList[i].Configuration = nil
	}

	return sourceList
}

func ParseSourceRead(sourceRead airbyteclient.SourceRead) *dataextractor.Source {
	configuration := (dataextractor.SourceConfiguration)(sourceRead.ConnectionConfiguration)
	return &dataextractor.Source{
		Name:               sourceRead.Name,
		SourceDefinitionId: sourceRead.SourceDefinitionId,
		SourceId:           sourceRead.SourceId,
		Configuration:      &configuration,
	}
}

const (
	typeKey        = "type"
	airbyteTypeKey = "airbyte_type"
	anyOfKey       = "anyOf"
	allOfKey       = "allOf"
	oneOfKey       = "oneOf"
	formatKey      = "format"
	propertiesKey  = "properties"
	itemsKey       = "items"
)

var (
	compositionKeys = map[string]bool{
		anyOfKey: true,
		allOfKey: true,
		oneOfKey: true,
	}
)

// We need to clean the JSON Schema to ensure we don't bump into the
// error scenarios on the table creation process in order to generate
// a table that Trino is able to read without a problem
func CleanJsonSchema(schema map[string]interface{}) map[string]interface{} {
	newSchema := map[string]interface{}{}

	schemaTypeInterface := schema[typeKey]
	schemaType := ""
	switch schemaTypeInterface.(type) {
	case []interface{}:
		schemaType = schemaTypeInterface.([]interface{})[1].(string)
	case string:
		schemaType = schema[typeKey].(string)
	default:
		for key := range compositionKeys {
			if _, exists := schema[key]; exists {
				schemaType = key
			}
		}
		if schemaType == "" {
			return newSchema
		}
	}

	switch schemaType {
	case "object":
		if properties, exists := schema[propertiesKey]; exists {
			newProperties := map[string]interface{}{}

			nestedProperties := properties.(map[string]interface{})
			for nestedColName, nestedValue := range nestedProperties {
				if nestedValue != nil && len(nestedValue.(map[string]interface{})) > 0 {
					value := CleanJsonSchema(nestedValue.(map[string]interface{}))
					if len(value) == 0 {
						continue
					}
					newProperties[nestedColName] = value
				}
			}
			if len(nestedProperties) > 0 {
				newSchema[typeKey] = schemaType
				newSchema[propertiesKey] = newProperties
			}
		}
	case "array":
		if items, exists := schema[itemsKey]; exists {
			newSchema[typeKey] = schemaType
			newSchema[itemsKey] = CleanJsonSchema(items.(map[string]interface{}))
		}
	default:
		if _, isCompositeKey := compositionKeys[schemaType]; isCompositeKey {
			if values, exists := schema[schemaType]; exists {
				newValues := []interface{}{}
				for _, value := range values.([]interface{}) {
					newValues = append(newValues, CleanJsonSchema(value.(map[string]interface{})))
				}
				newSchema[schemaType] = newValues
			}
		} else {
			if airbyteType, exists := schema[airbyteTypeKey]; exists {
				newSchema[airbyteTypeKey] = airbyteType
			}
			if format, exists := schema[formatKey]; exists {
				newSchema[formatKey] = format
			}
			newSchema[typeKey] = schemaType
		}
	}
	return newSchema
}

func ParseSyncCatalogToTables(syncCatalog *airbyteclient.AirbyteCatalog) []*dataextractor.Table {
	tables := make([]*dataextractor.Table, len(syncCatalog.Streams))
	for j, stream := range syncCatalog.Streams {
		supportedSyncModes := make([]string, len(stream.Stream.SupportedSyncModes))
		for k, syncMode := range stream.Stream.SupportedSyncModes {
			supportedSyncModes[k] = string(syncMode)
		}

		selectedCursorField := []string{}
		if stream.Config.CursorField != nil {
			selectedCursorField = *stream.Config.CursorField
		}

		tables[j] = &dataextractor.Table{
			Name:                    stream.Stream.Name,
			JsonSchema:              CleanJsonSchema(stream.Stream.JsonSchema),
			SupportedSyncModes:      supportedSyncModes,
			SourceDefinedCursor:     stream.Stream.SourceDefinedCursor,
			DefaultCursorField:      stream.Stream.DefaultCursorField,
			SourceDefinedPrimaryKey: stream.Stream.SourceDefinedPrimaryKey,
			SyncEnabled:             *stream.Config.Selected,
			Namespace:               stream.Stream.Namespace,
			SelectedCursorField:     &selectedCursorField,
		}
	}
	return tables
}

func ParseConnectionRead(connection *airbyteclient.ConnectionRead) *dataextractor.Connector {
	tables := ParseSyncCatalogToTables(&connection.SyncCatalog)
	var schedule *dataextractor.Schedule
	var basicSchedule *dataextractor.BasicSchedule
	var cronSchedule *dataextractor.CronSchedule
	if connection.ScheduleData != nil && connection.ScheduleData.BasicSchedule != nil {
		basicSchedule = &dataextractor.BasicSchedule{
			TimeUnit: string(connection.ScheduleData.BasicSchedule.TimeUnit),
			Units:    connection.ScheduleData.BasicSchedule.Units,
		}
	}

	if connection.ScheduleData != nil && connection.ScheduleData.Cron != nil {
		cronSchedule = &dataextractor.CronSchedule{
			Expression: string(connection.ScheduleData.Cron.CronExpression),
			TimeZone:   connection.ScheduleData.Cron.CronTimeZone,
		}
	}

	schedule = &dataextractor.Schedule{
		BasicSchedule: basicSchedule,
		CronSchedule:  cronSchedule,
	}

	sourceConfiguration := map[string]interface{}{}
	destinationConfiguration := map[string]interface{}{}
	if connection.Source != nil && connection.Destination != nil {
		sourceConfiguration = connection.Source.ConnectionConfiguration
		destinationConfiguration = connection.Destination.ConnectionConfiguration
	}

	connector := dataextractor.Connector{
		ID:                       (*uuid.UUID)(&connection.ConnectionId),
		Name:                     connection.Name,
		SourceId:                 (*uuid.UUID)(&connection.SourceId),
		ResourceRequirements:     (*dataextractor.ResourceRequirements)(connection.ResourceRequirements),
		Schedule:                 schedule,
		Tables:                   tables,
		SourceConfiguration:      sourceConfiguration,
		DestinationConfiguration: destinationConfiguration,
	}
	connector.WarehouseDatabase = connector.GetWarehouseDatabaseName()
	if connection.ScheduleType != nil {
		connector.ScheduleType = (string)(*connection.ScheduleType)
	}

	return &connector
}

func ParseListConnectionsForWorkspaceWithResponse(connections *airbyteclient.ConnectionReadList) []*dataextractor.Connector {
	connectors := make([]*dataextractor.Connector, len(connections.Connections))
	for i, connector := range connections.Connections {
		connectors[i] = ParseConnectionRead(&connector)
	}

	return connectors
}

func ParseConnectorResourceRequirementsToAirbyte(resources *dataextractor.ResourceRequirements) *airbyteclient.ResourceRequirements {
	if resources == nil {
		return nil
	}

	return &airbyteclient.ResourceRequirements{
		CpuLimit:      resources.CpuLimit,
		CpuRequest:    resources.CpuRequest,
		MemoryLimit:   resources.MemoryLimit,
		MemoryRequest: resources.MemoryRequest,
	}
}
func ParseConnectorScheduleToAirbyte(schedule *dataextractor.Schedule) (*airbyteclient.ConnectionSchedule, *airbyteclient.ConnectionScheduleData) {
	if schedule == nil {
		return nil, nil
	}

	var connectionSchedule *airbyteclient.ConnectionSchedule
	var basicSchedule *struct {
		TimeUnit airbyteclient.ConnectionScheduleDataBasicScheduleTimeUnit `json:"timeUnit"`
		Units    int64                                                     `json:"units"`
	} = nil
	var cronSchedule *struct {
		CronExpression string `json:"cronExpression"`
		CronTimeZone   string `json:"cronTimeZone"`
	} = nil
	if schedule.BasicSchedule != nil {
		connectionSchedule = &airbyteclient.ConnectionSchedule{
			TimeUnit: airbyteclient.ConnectionScheduleTimeUnit(schedule.BasicSchedule.TimeUnit),
			Units:    int64(schedule.BasicSchedule.Units),
		}

		basicSchedule = &struct {
			TimeUnit airbyteclient.ConnectionScheduleDataBasicScheduleTimeUnit `json:"timeUnit"`
			Units    int64                                                     `json:"units"`
		}{
			TimeUnit: airbyteclient.ConnectionScheduleDataBasicScheduleTimeUnit(schedule.BasicSchedule.TimeUnit),
			Units:    int64(schedule.BasicSchedule.Units),
		}
	} else if schedule.CronSchedule != nil {
		cronSchedule = &struct {
			CronExpression string `json:"cronExpression"`
			CronTimeZone   string `json:"cronTimeZone"`
		}{
			CronExpression: schedule.CronSchedule.Expression,
			CronTimeZone:   schedule.CronSchedule.TimeZone,
		}
	}

	return connectionSchedule, &airbyteclient.ConnectionScheduleData{
		BasicSchedule: basicSchedule,
		Cron:          cronSchedule,
	}
}

func ParseConnectorTablesToSyncCatalog(tables []*dataextractor.Table) airbyteclient.AirbyteCatalog {
	streams := make([]airbyteclient.AirbyteStreamAndConfiguration, len(tables))
	for i, table := range tables {
		syncModes := make([]airbyteclient.SyncMode, len(table.SupportedSyncModes))
		for j, syncMode := range table.SupportedSyncModes {
			syncModes[j] = airbyteclient.SyncMode(syncMode)
		}

		cursorField := []string{}
		if table.SelectedCursorField != nil {
			cursorField = *table.SelectedCursorField
		} else if len(table.DefaultCursorField) > 0 {
			cursorField = table.DefaultCursorField
		}

		destinationSyncMode := airbyteclient.Overwrite
		syncMode := airbyteclient.FullRefresh
		if len(syncMode) > 1 && len(cursorField) > 0 {
			destinationSyncMode = airbyteclient.Append
			syncMode = airbyteclient.Incremental
		}

		streams[i] = airbyteclient.AirbyteStreamAndConfiguration{
			Stream: airbyteclient.AirbyteStream{
				Name:                    table.Name,
				JsonSchema:              table.JsonSchema,
				SupportedSyncModes:      syncModes,
				SourceDefinedCursor:     table.SourceDefinedCursor,
				DefaultCursorField:      table.DefaultCursorField,
				SourceDefinedPrimaryKey: table.SourceDefinedPrimaryKey,
				Namespace:               table.Namespace,
			},
			Config: airbyteclient.AirbyteStreamConfiguration{
				SyncMode:            syncMode,
				CursorField:         &cursorField,
				DestinationSyncMode: destinationSyncMode,
				PrimaryKey:          &table.SourceDefinedPrimaryKey,
				AliasName:           &table.Name,
				Selected:            &table.SyncEnabled,
			},
		}
	}

	return airbyteclient.AirbyteCatalog{streams}
}

func GenerateCreateConnectionRequestBody(conn *dataextractor.Connector, destinationId *uuid.UUID) airbyteclient.WebBackendConnectionCreate {
	namespaceDefinition := airbyteclient.Customformat
	namespaceFormat := conn.GetWarehouseDatabaseName()

	body := airbyteclient.WebBackendConnectionCreate{
		SourceId:                     *conn.SourceId,
		DestinationId:                *destinationId,
		Name:                         &conn.Name,
		NamespaceDefinition:          &namespaceDefinition,
		NamespaceFormat:              &namespaceFormat,
		ScheduleType:                 airbyteclient.ConnectionScheduleType(conn.ScheduleType),
		Status:                       airbyteclient.Active,
		NonBreakingChangesPreference: airbyteclient.PropagateColumns,
	}

	body.SyncCatalog = ParseConnectorTablesToSyncCatalog(conn.Tables)

	connectionSchedule, connectionScheduleData := ParseConnectorScheduleToAirbyte(conn.Schedule)
	body.Schedule = connectionSchedule
	body.ScheduleData = connectionScheduleData

	body.ResourceRequirements = ParseConnectorResourceRequirementsToAirbyte(conn.ResourceRequirements)

	return body
}

func GenerateUpdateConnectionRequestBody(conn *dataextractor.Connector) airbyteclient.WebBackendConnectionUpdate {
	namespaceDefinition := airbyteclient.Customformat
	namespaceFormat := conn.GetWarehouseDatabaseName()

	body := airbyteclient.WebBackendConnectionUpdate{
		Name:                         &conn.Name,
		ConnectionId:                 *conn.ID,
		NamespaceDefinition:          &namespaceDefinition,
		NamespaceFormat:              &namespaceFormat,
		ScheduleType:                 airbyteclient.ConnectionScheduleType(conn.ScheduleType),
		Status:                       airbyteclient.Active,
		NonBreakingChangesPreference: airbyteclient.PropagateColumns,
	}

	body.SyncCatalog = ParseConnectorTablesToSyncCatalog(conn.Tables)

	connectionSchedule, connectionScheduleData := ParseConnectorScheduleToAirbyte(conn.Schedule)
	body.Schedule = connectionSchedule
	body.ScheduleData = connectionScheduleData

	body.ResourceRequirements = ParseConnectorResourceRequirementsToAirbyte(conn.ResourceRequirements)

	return body
}
