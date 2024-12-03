package ingestions

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	dataextractor "github.com/ubix/dripper/manager/repositories/data-extractor"
)

type EndpointSet struct {
	CreateWorkspaceEndpoint endpoint.Endpoint
	ListWorkspacesEndpoint  endpoint.Endpoint
	GetWorkspaceEndpoint    endpoint.Endpoint

	ListSourcesDefinitionsEndpoint                 endpoint.Endpoint
	GetSourceDefinitionConfigurationEndpoint       endpoint.Endpoint
	CreateSourceFromDefinitionAsyncRequestEndpoint endpoint.Endpoint
	CreateSourceFromDefinitionEndpoint             endpoint.Endpoint
	GetCreateSourceFromDefinitionStatusEndpoint    endpoint.Endpoint

	ListSourcesEndpoint               endpoint.Endpoint
	GetSourceEndpoint                 endpoint.Endpoint
	UpdateSourceAsyncRequestEndpoint  endpoint.Endpoint
	UpdateSourceEndpoint              endpoint.Endpoint
	GetUpdateSourceStatusEndpoint     endpoint.Endpoint
	DeleteSourceEndpoint              endpoint.Endpoint
	CreateConnectorFromSourceEndpoint endpoint.Endpoint

	GetSourceTablesAsyncRequestEndpoint endpoint.Endpoint
	GetSourceTablesEndpoint             endpoint.Endpoint
	GetGetSourceTablesStatusEndpoint    endpoint.Endpoint

	ListConnectorsEndpoint  endpoint.Endpoint
	GetConnectorEndpoint    endpoint.Endpoint
	UpdateConnectorEndpoint endpoint.Endpoint
	DeleteConnectorEndpoint endpoint.Endpoint

	TriggerConnectorJobEndpoint endpoint.Endpoint
	ListConnectorJobsEndpoint   endpoint.Endpoint
	GetConnectorJobEndpoint     endpoint.Endpoint
	GetConnectorJobLogsEndpoint endpoint.Endpoint

	WebhookEndpoint          endpoint.Endpoint
	TriggerProfilingEndpoint endpoint.Endpoint
}

func NewEndpointSet(svc ConnectorService) *EndpointSet {
	return &EndpointSet{
		CreateWorkspaceEndpoint: makeCreateWorkspaceEndpoint(svc),
		ListWorkspacesEndpoint:  makeListWorkspacesEndpoint(svc),
		GetWorkspaceEndpoint:    makeGetWorkspaceEndpoint(svc),

		ListSourcesDefinitionsEndpoint:                 makeListSourcesDefinitionsEndpoint(svc),
		GetSourceDefinitionConfigurationEndpoint:       makeGetSourceDefinitionConfigurationEndpoint(svc),
		CreateSourceFromDefinitionAsyncRequestEndpoint: makeCreateSourceFromDefinitionAsyncRequestEndpoint(svc),
		CreateSourceFromDefinitionEndpoint:             makeCreateSourceFromDefinition(svc),
		GetCreateSourceFromDefinitionStatusEndpoint:    makeGetCreateSourceFromDefinitionStatus(svc),

		ListSourcesEndpoint:              makeListSourcesEndpoint(svc),
		GetSourceEndpoint:                makeGetSourceEndpoint(svc),
		UpdateSourceEndpoint:             makeUpdateSourceEndpoint(svc),
		UpdateSourceAsyncRequestEndpoint: makeUpdateSourceAsyncRequestEndpoint(svc),
		GetUpdateSourceStatusEndpoint:    makeGetUpdateSourceStatusEndpoint(svc),
		DeleteSourceEndpoint:             makeDeleteSourceEndpoint(svc),

		GetSourceTablesAsyncRequestEndpoint: makeGetSourceTablesAsyncRequestEndpoint(svc),
		GetSourceTablesEndpoint:             makeGetSourceTablesEndpoint(svc),
		GetGetSourceTablesStatusEndpoint:    makeGetGetSourceTablesStatusEndpoint(svc),

		CreateConnectorFromSourceEndpoint: makeCreateConnectorFromSourceEndpoint(svc),
		ListConnectorsEndpoint:            makeListConnectorsEndpoint(svc),
		GetConnectorEndpoint:              makeGetConnectorEndpoint(svc),
		UpdateConnectorEndpoint:           makeUpdateConnectorEndpoint(svc),
		DeleteConnectorEndpoint:           makeDeleteConnectorEndpoint(svc),
		TriggerConnectorJobEndpoint:       makeTriggerConnectorJobEndpoint(svc),

		ListConnectorJobsEndpoint:   makeListConnectorJobsEndpoint(svc),
		GetConnectorJobEndpoint:     makeGetConnectorJobEndpoint(svc),
		GetConnectorJobLogsEndpoint: makeGetConnectorJobLogsEndpoint(svc),

		WebhookEndpoint:          makeWebhookEndpoint(svc),
		TriggerProfilingEndpoint: makeTriggerProfilingEndpoint(svc),
	}
}

func makeCreateWorkspaceEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createWorkspaceRequest)
		return svc.CreateWorkspace(ctx, req.Tenant, req.Bucket)
	}
}

func makeListWorkspacesEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listWorkspacesRequest)
		pag := dataextractor.Filters{
			NameContains: req.NameContains,
			Limit:        req.Limit,
			Offset:       req.Offset,
		}

		return svc.ListWorkspaces(ctx, pag)
	}
}

func makeGetWorkspaceEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getWorkspaceRequest)
		return svc.GetWorkspace(ctx, req.workspaceId)
	}
}

func makeListSourcesDefinitionsEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listSourcesDefinitionsRequest)
		return svc.ListSourcesDefinitions(ctx, req.workspaceId)
	}
}

func makeGetSourceDefinitionConfigurationEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getSourceDefinitionSpecificationRequest)
		return svc.GetSourceDefinitionConfiguration(ctx, req.workspaceId, req.SourceDefinitionId)
	}
}

func makeCreateSourceFromDefinitionAsyncRequestEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createSourceFromDefinitionAsyncRequest)

		ticketId, err := svc.CreateSourceFromDefinitionAsync(ctx, req.WorkspaceId, req.Source)

		if err != nil {
			return nil, err
		}

		return &asyncOperationResponse{
			TicketId: ticketId,
		}, nil
	}
}

func makeCreateSourceFromDefinition(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*createSourceFromDefinitionRequestMessage)

		err := svc.CreateSourceFromDefinition(ctx, req.body.WorkspaceId, req.body.Source, req.ticketId)
		if err != nil {
			return nil, err
		}

		return &createSourceFromDefinitionResponseMessage{
			SourceID: &req.body.Source.SourceId,
		}, nil
	}
}

func makeGetCreateSourceFromDefinitionStatus(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*getStatusRequest)

		status, err := svc.GetCreateSourceFromDefinitionStatus(ctx, req.ticketId)
		if err != nil {
			return nil, err
		}

		return status, nil
	}
}

func makeListSourcesEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listSourcesRequest)
		return svc.ListSources(ctx, req.workspaceId)
	}
}

func makeGetSourceEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getSourceRequest)
		return svc.GetSource(ctx, req.sourceId)
	}
}

func makeGetSourceTablesAsyncRequestEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getSourceTablesAsyncRequest)

		ticketId, err := svc.GetSourceTablesAsync(ctx, req)
		if err != nil {
			return nil, err
		}

		return &asyncOperationResponse{
			TicketId: ticketId,
		}, nil
	}
}

func makeGetSourceTablesEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*getSourceTablesRequestMessage)
		tables, err := svc.GetSourceTables(ctx, req.sourceId, req.ticketId)
		if err != nil {
			return nil, err
		}

		return &getSourceTablesResponseMessage{tables}, nil
	}
}

func makeGetGetSourceTablesStatusEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*getStatusRequest)

		status, err := svc.GetGetSourceTablesStatus(ctx, req.ticketId)
		if err != nil {
			return nil, err
		}

		return status, nil
	}
}

// TODO: Check api call on the browser's console
func makeUpdateSourceAsyncRequestEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateSourceAsyncRequest)

		ticketId, err := svc.UpdateSourceAsync(ctx, req)

		if err != nil {
			return nil, err
		}

		return &asyncOperationResponse{
			TicketId: ticketId,
		}, nil
	}
}

func makeUpdateSourceEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*updateSourceAsyncRequestMessage)

		res, err := svc.UpdateSource(ctx, req.source, req.ticketId)
		if err != nil {
			return nil, err
		}

		return &updateSourceAsyncResponseMessage{
			Source: res,
		}, nil
	}
}

func makeGetUpdateSourceStatusEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*getStatusRequest)

		status, err := svc.GetUpdateSourceStatus(ctx, req.ticketId)
		if err != nil {
			return nil, err
		}

		return status, nil
	}
}

func makeDeleteSourceEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteSourceRequest)
		return nil, svc.DeleteSource(ctx, req.sourceId)
	}
}

func makeCreateConnectorFromSourceEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createConnectorFromSourceRequest)

		res, err := svc.CreateConnectorFromSource(ctx, req.WorkspaceId, req.connector)
		if err != nil {
			return nil, err
		}

		return res, nil
	}
}

func makeUpdateConnectorEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateConnectorRequest)

		res, err := svc.UpdateConnector(ctx, req)
		if err != nil {
			return nil, err
		}

		return res, nil
	}
}

func makeListConnectorsEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listConnectorsRequest)
		pag := &dataextractor.Filters{
			WorkspaceId: req.WorkspaceId,
			Limit:       req.Limit,
			Offset:      req.Offset,
		}

		connectors, err := svc.ListConnectors(ctx, pag)
		if err != nil {
			return nil, err
		}

		res := listConnectorsResponse(connectors)
		return res, nil
	}
}

func makeGetConnectorEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getConnectorRequest)
		conn, err := svc.GetConnector(ctx, req.connectionId)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

func makeDeleteConnectorEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteConnectorRequest)
		err := svc.DeleteConnector(ctx, req.connectionId)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func makeTriggerConnectorJobEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(triggerJobRequest)
		job, err := svc.TriggerJob(ctx, req.ConnectionId)
		if err != nil {
			return nil, err
		}
		res := triggerJobResponse(job)
		return res, nil
	}
}

func makeGetConnectorJobEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getConnectorJobRequest)
		job, err := svc.GetConnectorJob(ctx, req.JobId)
		if err != nil {
			return nil, err
		}
		res := getConnectorJobResponse(job)
		return res, nil
	}
}

func makeGetConnectorJobLogsEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getConnectorJobRequest)
		logs, err := svc.GetConnectorJobLogs(ctx, req.JobId)
		if err != nil {
			return nil, err
		}
		return logs, nil
	}
}

func makeListConnectorJobsEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listJobsRequest)
		pag := &dataextractor.Filters{
			Limit:          req.Limit,
			Offset:         req.Offset,
			JobType:        req.JobType,
			Status:         req.Status,
			CreatedAtStart: req.CreatedAtStart,
			CreatedAtEnd:   req.CreatedAtEnd,
			UpdatedAtStart: req.UpdatedAtStart,
			UpdatedAtEnd:   req.UpdatedAtEnd,
			OrderBy:        req.OrderBy,
		}

		jobs, err := svc.ListConnectorJobs(ctx, req.ConnectorId, pag)
		if err != nil {
			return nil, err
		}

		res := listJobsResponse(jobs)
		return res, nil
	}
}

func makeTriggerJobEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(triggerJobRequest)
		job, err := svc.TriggerJob(ctx, req.ConnectionId)
		if err != nil {
			return nil, err
		}
		res := triggerJobResponse(job)
		return res, nil
	}
}

func makeWebhookEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(webhookRequest)
		connectionId := req.Data.Connection.Id
		workspaceId := req.Data.Workspace.Id

		if connectionId == nil {
			return nil, nil
		}

		err := svc.SendWebhookToKafka(ctx, *connectionId, *workspaceId)
		if err != nil {
			return nil, err
		}

		return webhookResponse{Message: "Webhook processed and sent to Kafka successfully"}, nil
	}
}

func makeTriggerProfilingEndpoint(svc ConnectorService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*triggerProfilingRequestMessage)

		err := svc.TriggerProfiling(ctx, req.ConnectionId, req.WorkspaceId)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}

type triggerProfilingRequestMessage struct {
	WorkspaceId  uuid.UUID
	ConnectionId uuid.UUID
}

type webhookRequest struct {
	Data webhookData `json:"data"`
}

type webhookData struct {
	Workspace  workspace  `json:"workspace"`
	Connection connection `json:"connection"`
}

type createWorkspaceRequest struct {
	Tenant string `json:"tenant"`
	Bucket string `json:"bucket"`
}

type getWorkspaceRequest struct {
	workspaceId uuid.UUID `json:"workspaceId"`
}

type listSourcesRequest struct {
	workspaceId uuid.UUID `json:"workspaceId"`
}

type workspace struct {
	Id *uuid.UUID `json:"id"`
}

type connection struct {
	Id *uuid.UUID `json:"id"`
}

type webhookResponse struct {
	Message string `json:"message"`
}

type successfulSyncPayload struct {
	ConnectionId *uuid.UUID `json:"connectionId"`
	WorkspaceId  *uuid.UUID `json:"workspaceId"`
}

// Request/Response schemas
type asyncOperationResponse struct {
	TicketId *uuid.UUID `json:"ticketId"`
}

type getStatusRequest struct {
	ticketId *uuid.UUID
}

type createConnectorRequest struct {
	Name                string                            `json:"name"`
	SourceID            uuid.UUID                         `json:"sourceId"`
	SourceConfiguration dataextractor.SourceConfiguration `json:"sourceConfiguration"`
}

type createConnectorAsyncResponse struct {
	connector *dataextractor.Connector
	ticketId  *uuid.UUID
}

type getConnectorRequest struct {
	connectionId uuid.UUID
}

type createSourceFromDefinitionAsyncRequestBody *dataextractor.Source

type createSourceFromDefinitionAsyncRequest struct {
	Source      *dataextractor.Source
	WorkspaceId uuid.UUID
}

type createSourceFromDefinitionRequestMessage struct {
	body     createSourceFromDefinitionAsyncRequest
	ticketId *uuid.UUID
}

type createSourceFromDefinitionResponseMessage struct {
	SourceID *uuid.UUID `json:"sourceId"`
}

type storeCreateSourceFromDefinitionResponseMessage struct {
	sourceID *uuid.UUID
	ticketId *uuid.UUID
}

type createSourceFromDefinitionStatusResponse *CreateSourceStatus

type updateConnectorRequest *dataextractor.Connector

type createConnectorFromSourceRequestBody *dataextractor.Connector

type createConnectorFromSourceRequest struct {
	connector   *dataextractor.Connector
	WorkspaceId uuid.UUID
}

type createConnectorFromSourceResponse *dataextractor.Connector

type getSourceTablesAsyncRequest *uuid.UUID

type getSourceTablesRequestMessage struct {
	sourceId getSourceTablesAsyncRequest
	ticketId *uuid.UUID
}

type getSourceTablesResponseMessage struct {
	tables []*dataextractor.Table
}

type storeGetSourceTablesResponseMessage struct {
	tables   []*dataextractor.Table
	ticketId *uuid.UUID
}

type listWorkspacesRequest struct {
	NameContains string
	Limit        int
	Offset       int
}

type listConnectorsRequest struct {
	WorkspaceId uuid.UUID
	Limit       int
	Offset      int
}

type deleteConnectorRequest struct {
	connectionId uuid.UUID
}

type listSourcesDefinitionsRequest struct {
	workspaceId uuid.UUID
}

type getSourceRequest struct {
	sourceId uuid.UUID
}

type deleteSourceRequest struct {
	sourceId uuid.UUID
}

type updateSourceAsyncRequest *dataextractor.Source

type updateSourceAsyncRequestMessage struct {
	source   updateSourceAsyncRequest
	ticketId *uuid.UUID
}

type updateSourceAsyncResponseMessage struct {
	Source *dataextractor.Source
}

type getSourceDefinitionSpecificationRequest struct {
	SourceDefinitionId string `json:"sourcedefinitionid"`
	workspaceId        uuid.UUID
}

type getCreateConnectorResponsRequest struct {
	ticketId *uuid.UUID
}

type createConnectorResponse *dataextractor.Connector

type listConnectorsResponse []*dataextractor.Connector

// Request/Response schemas
type listJobsRequest struct {
	ConnectorId    *uuid.UUID
	Limit          int
	Offset         int
	JobType        string
	WorkspaceIds   string
	Status         string
	CreatedAtStart string
	CreatedAtEnd   string
	UpdatedAtStart string
	UpdatedAtEnd   string
	OrderBy        string
}

type getConnectorJobRequest struct {
	ConnectorId *uuid.UUID
	JobId       int64
}

type triggerJobRequest struct {
	ConnectionId uuid.UUID
}

type listJobsResponse []*dataextractor.Job

type getConnectorJobResponse *dataextractor.Job

type triggerJobResponse *dataextractor.Job
