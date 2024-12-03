package ingestions

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/ubix/dripper/manager/constants"
	"github.com/ubix/dripper/manager/repositories/kvstore"

	datacatalog "github.com/ubix/dripper/manager/repositories/data-catalog"
	dataextractor "github.com/ubix/dripper/manager/repositories/data-extractor"
)

type Status string

const (
	Pending  Status = "pending"
	Finished Status = "finished"
	Failed   Status = "failed"
)

var (
	statusMap = map[string]Status{
		"pending":  Pending,
		"finished": Finished,
		"failed":   Failed,
	}
)

type CreateConnectorStatus struct {
	ConnectorId *uuid.UUID `json:"connectorId"`
	Status      Status     `json:"status"`
}

type CreateSourceStatus struct {
	SourceId *uuid.UUID `json:"sourceId"`
	Status   Status     `json:"status"`
	Error    string     `json:"error"`
}

type GetSourceTablesStatus struct {
	Tables []*dataextractor.Table `json:"tables"`
	Status Status                 `json:"status"`
	Error  string                 `json:"error"`
}

type UpdateSourceStatus struct {
	Source *dataextractor.Source `json:"source"`
	Status Status                `json:"status"`
	Error  string                `json:"error"`
}

// Service Interface
type ConnectorService interface {
	ListWorkspaces(ctx context.Context, pag dataextractor.Filters) ([]*dataextractor.Workspace, error)
	GetWorkspace(ctx context.Context, workspaceId uuid.UUID) (*dataextractor.Workspace, error)
	CreateWorkspace(ctx context.Context, tenant, bucket string) (*dataextractor.Workspace, error)
	ListSourcesDefinitions(ctx context.Context, workspaceId uuid.UUID) ([]*dataextractor.SourceDefinition, error)
	GetSourceDefinitionConfiguration(ctx context.Context, workspaceId uuid.UUID, sourceDefinitionId string) (*dataextractor.SourceDefinitionConfiguration, error)
	GetConnector(ctx context.Context, connectorId uuid.UUID) (*dataextractor.Connector, error)
	CreateSourceFromDefinitionAsync(ctx context.Context, workspaceId uuid.UUID, source *dataextractor.Source) (*uuid.UUID, error)
	CreateSourceFromDefinition(ctx context.Context, workspaceId uuid.UUID, source *dataextractor.Source, ticketId *uuid.UUID) error
	GetCreateSourceFromDefinitionStatus(ctx context.Context, ticketId *uuid.UUID) (*CreateSourceStatus, error)
	CreateConnectorFromSource(ctx context.Context, workspaceId uuid.UUID, connector *dataextractor.Connector) (*dataextractor.Connector, error)
	UpdateConnector(ctx context.Context, connector *dataextractor.Connector) (*dataextractor.Connector, error)
	GetSource(ctx context.Context, sourceId uuid.UUID) (*dataextractor.Source, error)
	DeleteSource(ctx context.Context, sourceId uuid.UUID) error
	UpdateSourceAsync(ctx context.Context, source *dataextractor.Source) (*uuid.UUID, error)
	UpdateSource(ctx context.Context, source *dataextractor.Source, ticketId *uuid.UUID) (*dataextractor.Source, error)
	GetUpdateSourceStatus(ctx context.Context, ticketId *uuid.UUID) (*UpdateSourceStatus, error)
	ListSources(ctx context.Context, workspaceId uuid.UUID) ([]*dataextractor.Source, error)
	GetSourceTablesAsync(ctx context.Context, sourceId *uuid.UUID) (*uuid.UUID, error)
	GetSourceTables(ctx context.Context, sourceId *uuid.UUID, ticketId *uuid.UUID) ([]*dataextractor.Table, error)
	GetGetSourceTablesStatus(ctx context.Context, ticketId *uuid.UUID) (*GetSourceTablesStatus, error)
	ListConnectors(ctx context.Context, pag *dataextractor.Filters) ([]*dataextractor.Connector, error)
	DeleteConnector(ctx context.Context, connectorId uuid.UUID) error
	ListConnectorJobs(ctx context.Context, connectorId *uuid.UUID, pag *dataextractor.Filters) ([]*dataextractor.Job, error)
	GetConnectorJob(ctx context.Context, jobId int64) (*dataextractor.Job, error)
	GetConnectorJobLogs(ctx context.Context, jobId int64) ([]string, error)

	TriggerJob(ctx context.Context, connectionId uuid.UUID) (*dataextractor.Job, error)

	SendWebhookToKafka(ctx context.Context, connectionId uuid.UUID, workspaceId uuid.UUID) error
	TriggerProfiling(ctx context.Context, connectionId uuid.UUID, workspaceId uuid.UUID) error
}

// Service Implementation
type connectorService struct {
	dataExtractor dataextractor.DataExtractor
	dataCatalog   datacatalog.DataCatalog
	publisher     message.Publisher
	store         kvstore.KVStore
}

func NewConnectorService(dataextractor dataextractor.DataExtractor, datacatalog datacatalog.DataCatalog, publisher message.Publisher, store kvstore.KVStore) ConnectorService {
	return &connectorService{
		dataExtractor: dataextractor,
		dataCatalog:   datacatalog,
		publisher:     publisher,
		store:         store,
	}
}

func (s *connectorService) ListWorkspaces(ctx context.Context, pag dataextractor.Filters) ([]*dataextractor.Workspace, error) {
	return s.dataExtractor.ListWorkspaces(ctx, pag)
}

func (s *connectorService) GetWorkspace(ctx context.Context, workspaceId uuid.UUID) (*dataextractor.Workspace, error) {
	return s.dataExtractor.GetWorkspace(ctx, workspaceId)
}

func (s *connectorService) CreateWorkspace(ctx context.Context, tenant, bucket string) (*dataextractor.Workspace, error) {
	return s.dataExtractor.CreateWorkspace(ctx, tenant, bucket)
}

func (s *connectorService) ListSourcesDefinitions(ctx context.Context, workspaceId uuid.UUID) ([]*dataextractor.SourceDefinition, error) {
	return s.dataExtractor.ListSourcesDefinitions(ctx, workspaceId)
}

func (s *connectorService) GetSourceDefinitionConfiguration(ctx context.Context, workspaceId uuid.UUID, id string) (*dataextractor.SourceDefinitionConfiguration, error) {
	return s.dataExtractor.GetSourceDefinitionConfiguration(ctx, workspaceId, id)
}

func (s *connectorService) CreateSourceFromDefinitionAsync(ctx context.Context, workspaceId uuid.UUID, source *dataextractor.Source) (*uuid.UUID, error) {
	jsonBody := createSourceFromDefinitionAsyncRequest{
		WorkspaceId: workspaceId,
		Source:      source,
	}
	payload, err := json.Marshal(jsonBody)
	if err != nil {
		return nil, err
	}

	ticketId, err := uuid.Parse(watermill.NewUUID())
	if err != nil {
		return nil, err
	}

	msg := message.NewMessage(ticketId.String(), payload)
	msg.SetContext(ctx)

	err = s.publisher.Publish(constants.BrokerTopics.CreateSourceRequests, msg)
	if err != nil {
		return nil, err
	}

	err = s.store.Set(ctx, ticketId, &CreateSourceStatus{
		Status: Pending,
	})
	if err != nil {
		return nil, err
	}

	return &ticketId, nil
}

func (s *connectorService) CreateSourceFromDefinition(ctx context.Context, workspaceId uuid.UUID, source *dataextractor.Source, ticketId *uuid.UUID) error {
	err := s.dataExtractor.CreateSourceFromDefinition(ctx, workspaceId, source)

	if err != nil {
		s.store.Set(ctx, *ticketId, &CreateSourceStatus{
			Status: Failed,
			Error:  err.Error(),
		})
		return err
	}

	s.store.Set(ctx, *ticketId, &CreateSourceStatus{
		Status:   Finished,
		SourceId: &source.SourceId,
	})

	return nil
}

func (s *connectorService) GetCreateSourceFromDefinitionStatus(ctx context.Context, ticketId *uuid.UUID) (*CreateSourceStatus, error) {
	value, err := s.store.Get(ctx, *ticketId)
	if err != nil {
		return nil, err
	}

	errorString := ""
	valueMap := value.(map[string]interface{})
	var sourceId *uuid.UUID
	if valueMap["sourceId"] != nil {
		sourceIdParsed, err := uuid.Parse(valueMap["sourceId"].(string))
		if err != nil {
			return nil, err
		}

		sourceId = &sourceIdParsed
	} else if valueMap["error"] != nil {
		errorString = valueMap["error"].(string)
	}

	statusString := valueMap["status"].(string)
	status := statusMap[statusString]

	return &CreateSourceStatus{
		SourceId: sourceId,
		Status:   status,
		Error:    errorString,
	}, nil
}

func (s *connectorService) ListSources(ctx context.Context, workspaceId uuid.UUID) ([]*dataextractor.Source, error) {
	return s.dataExtractor.ListSources(ctx, workspaceId)
}

func (s *connectorService) GetSource(ctx context.Context, sourceId uuid.UUID) (*dataextractor.Source, error) {
	return s.dataExtractor.GetSource(ctx, sourceId)
}

func (s *connectorService) UpdateSourceAsync(ctx context.Context, source *dataextractor.Source) (*uuid.UUID, error) {
	payload, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}

	ticketId, err := uuid.Parse(watermill.NewUUID())
	if err != nil {
		return nil, err
	}

	msg := message.NewMessage(ticketId.String(), payload)
	msg.SetContext(ctx)

	err = s.publisher.Publish(constants.BrokerTopics.UpdateSourceRequests, msg)
	if err != nil {
		return nil, err
	}

	err = s.store.Set(ctx, ticketId, &UpdateSourceStatus{
		Status: Pending,
	})
	if err != nil {
		return nil, err
	}

	return &ticketId, nil
}

func (s *connectorService) UpdateSource(ctx context.Context, source *dataextractor.Source, ticketId *uuid.UUID) (*dataextractor.Source, error) {
	source, err := s.dataExtractor.UpdateSource(ctx, source)

	if err != nil {
		s.store.Set(ctx, *ticketId, &UpdateSourceStatus{
			Status: Failed,
			Error:  err.Error(),
		})
		return nil, err
	}

	s.store.Set(ctx, *ticketId, &UpdateSourceStatus{
		Status: Finished,
		Source: source,
	})

	return source, nil
}

func (s *connectorService) GetUpdateSourceStatus(ctx context.Context, ticketId *uuid.UUID) (*UpdateSourceStatus, error) {
	value, err := s.store.Get(ctx, *ticketId)
	if err != nil {
		return nil, err
	}

	return ParseUpdateSourceStatusFromKafka(value)
}

func (s *connectorService) DeleteSource(ctx context.Context, sourceId uuid.UUID) error {
	return s.dataExtractor.DeleteSource(ctx, sourceId)
}

func (s *connectorService) GetSourceTablesAsync(ctx context.Context, sourceId *uuid.UUID) (*uuid.UUID, error) {
	// Check if source exists before creating async job
	source, err := s.dataExtractor.GetSource(ctx, *sourceId)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(source.SourceId)
	if err != nil {
		return nil, err
	}

	ticketId, err := uuid.Parse(watermill.NewUUID())
	if err != nil {
		return nil, err
	}

	msg := message.NewMessage(ticketId.String(), payload)
	msg.SetContext(ctx)

	err = s.publisher.Publish(constants.BrokerTopics.GetSourceTablesRequests, msg)
	if err != nil {
		return nil, err
	}

	err = s.store.Set(ctx, ticketId, &GetSourceTablesStatus{
		Status: Pending,
	})
	if err != nil {
		return nil, err
	}

	return &ticketId, nil
}

func (s *connectorService) GetSourceTables(ctx context.Context, sourceId *uuid.UUID, ticketId *uuid.UUID) ([]*dataextractor.Table, error) {
	// Check if source still exists before trying to get tables
	source, err := s.dataExtractor.GetSource(ctx, *sourceId)
	if err != nil {
		return nil, err
	}

	tables, err := s.dataExtractor.GetSourceTables(ctx, source.SourceId)
	if err != nil {
		s.store.Set(ctx, *ticketId, &GetSourceTablesStatus{
			Status: Failed,
			Error:  err.Error(),
		})
		return nil, err
	}

	s.store.Set(ctx, *ticketId, &GetSourceTablesStatus{
		Status: Finished,
		Tables: tables,
	})

	return tables, nil
}

func (s *connectorService) GetGetSourceTablesStatus(ctx context.Context, ticketId *uuid.UUID) (*GetSourceTablesStatus, error) {
	value, err := s.store.Get(ctx, *ticketId)
	if err != nil {
		return nil, err
	}
	message, err := ParseGetSourceTablesStatusMessage(value)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *connectorService) CreateConnectorFromSource(ctx context.Context, workspaceId uuid.UUID, connector *dataextractor.Connector) (*dataextractor.Connector, error) {
	connector, err := s.dataExtractor.CreateConnectorFromSource(ctx, workspaceId, connector)
	if err != nil {
		return nil, err
	}

	err = s.dataCatalog.CreateDatabase(ctx, connector)
	if err != nil {
		s.dataExtractor.DeleteConnector(ctx, *connector.ID)
		return nil, err
	}

	err = s.dataCatalog.UpdateTables(ctx, connector)
	if err != nil {
		s.dataCatalog.DeleteDatabase(ctx, connector)
		s.dataExtractor.DeleteConnector(ctx, *connector.ID)
		return nil, err
	}

	return connector, nil
}

func (s *connectorService) UpdateConnector(ctx context.Context, connector *dataextractor.Connector) (*dataextractor.Connector, error) {
	oldConnector, err := s.dataExtractor.GetConnector(ctx, *connector.ID)
	if err != nil {
		return nil, err
	}

	oldConnector.Name = connector.Name
	oldConnector.ScheduleType = connector.ScheduleType
	oldConnector.Schedule = connector.Schedule
	oldConnector.Tables = connector.Tables

	connector, err = s.dataExtractor.UpdateConnector(ctx, oldConnector)
	if err != nil {
		return nil, err
	}

	err = s.dataCatalog.UpdateTables(ctx, connector)
	if err != nil {
		return nil, err
	}

	return connector, nil
}

func (s *connectorService) GetConnector(ctx context.Context, connectorId uuid.UUID) (*dataextractor.Connector, error) {
	return s.dataExtractor.GetConnector(ctx, connectorId)
}

func (s *connectorService) ListConnectors(ctx context.Context, pag *dataextractor.Filters) ([]*dataextractor.Connector, error) {
	return s.dataExtractor.ListConnectors(ctx, pag)
}

func (s *connectorService) DeleteConnector(ctx context.Context, connectionId uuid.UUID) error {
	connector, err := s.dataExtractor.GetConnector(ctx, connectionId)
	if err != nil {
		return err
	}

	err = s.dataCatalog.DeleteDatabase(ctx, connector)
	if err != nil {
		return err
	}

	return s.dataExtractor.DeleteConnector(ctx, connectionId)
}

func (s *connectorService) TriggerJob(ctx context.Context, connectionId uuid.UUID) (*dataextractor.Job, error) {
	return s.dataExtractor.TriggerConnectorJob(ctx, connectionId)
}

func (s *connectorService) GetConnectorJob(ctx context.Context, jobId int64) (*dataextractor.Job, error) {
	return s.dataExtractor.GetConnectorJobDetails(ctx, jobId)
}

func (s *connectorService) ListConnectorJobs(ctx context.Context, connectorId *uuid.UUID, pag *dataextractor.Filters) ([]*dataextractor.Job, error) {
	return s.dataExtractor.ListConnectorJobs(ctx, connectorId, pag)
}

func (s *connectorService) GetConnectorJobLogs(ctx context.Context, jobId int64) ([]string, error) {
	return s.dataExtractor.GetConnectorJobLogs(ctx, jobId)
}

func (s *connectorService) SendWebhookToKafka(ctx context.Context, connectionId uuid.UUID, workspaceId uuid.UUID) error {
	kafkaPayload, err := json.Marshal(map[string]uuid.UUID{
		"workspaceId":  workspaceId,
		"connectionId": connectionId,
	})
	if err != nil {
		return err
	}

	msg := message.NewMessage(uuid.New().String(), kafkaPayload)

	err = s.publisher.Publish(constants.BrokerTopics.SuccessfulSyncNotificationsRequests, msg)
	if err != nil {
		return err
	}

	// Send message but try to refresh partitions right away to speedup process
	connector, err := s.dataExtractor.GetConnector(ctx, connectionId)
	if err != nil {
		return err
	}

	workspace, err := s.dataExtractor.GetWorkspace(ctx, workspaceId)
	if err != nil {
		return err
	}

	err = s.dataCatalog.RefreshTablesPartitions(ctx, connector, workspace.Name)
	if err != nil {
		return err
	}

	return nil
}

func (s *connectorService) TriggerProfiling(ctx context.Context, connectionId uuid.UUID, workspaceId uuid.UUID) error {
	connector, err := s.dataExtractor.GetConnector(ctx, connectionId)
	if err != nil {
		return err
	}

	workspace, err := s.dataExtractor.GetWorkspace(ctx, workspaceId)
	if err != nil {
		return err
	}

	err = s.dataCatalog.RefreshTablesPartitions(ctx, connector, workspace.Name)
	if err != nil {
		return err
	}

	err = s.dataCatalog.ProfileTables(ctx, connector)
	if err != nil {
		return err
	}

	return nil
}
