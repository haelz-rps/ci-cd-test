package ingestions

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	kitlog "github.com/go-kit/log"
	"github.com/google/uuid"
	dataextractor "github.com/ubix/dripper/manager/repositories/data-extractor"
)

func NewHTTPHandler(endpointSet *EndpointSet, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	// Workspaces

	createWorkspaceHandler := kithttp.NewServer(
		endpointSet.CreateWorkspaceEndpoint,
		decodeCreateWorkspaceRequest,
		encodeResponse(http.StatusCreated),
		opts...,
	)

	listWorkspacesHandler := kithttp.NewServer(
		endpointSet.ListWorkspacesEndpoint,
		decodeListWorkspacesRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	getWorkspaceHandler := kithttp.NewServer(
		endpointSet.GetWorkspaceEndpoint,
		decodeGetWorkspaceRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	//Source Definitions

	listSourcesDefinitionsHandler := kithttp.NewServer(
		endpointSet.ListSourcesDefinitionsEndpoint,
		decodeListSourcesDefinitionsRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	getSourceDefinitionConfigurationHandler := kithttp.NewServer(
		endpointSet.GetSourceDefinitionConfigurationEndpoint,
		decodeGetSourceDefinitionConfigurationRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	createSourceFromDefinitionHandler := kithttp.NewServer(
		endpointSet.CreateSourceFromDefinitionAsyncRequestEndpoint,
		decodeCreateSourceFromDefinitionAsyncRequest,
		encodeResponse(http.StatusAccepted),
		opts...,
	)

	getCreateSourceFromDefinitionStatus := kithttp.NewServer(
		endpointSet.GetCreateSourceFromDefinitionStatusEndpoint,
		decodeGetStatusRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	getSourceTablesHandler := kithttp.NewServer(
		endpointSet.GetSourceTablesAsyncRequestEndpoint,
		decodeGetSourceTablesAsyncRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	getGetSourceTablesStatus := kithttp.NewServer(
		endpointSet.GetGetSourceTablesStatusEndpoint,
		decodeGetStatusRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	createConnectorFromSourceHandler := kithttp.NewServer(
		endpointSet.CreateConnectorFromSourceEndpoint,
		decodeCreateConnectorFromSourceRequest,
		encodeResponse(http.StatusCreated),
		opts...,
	)

	// Source

	listSourcesHandler := kithttp.NewServer(
		endpointSet.ListSourcesEndpoint,
		decodeListSourcesRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	getSourceHandler := kithttp.NewServer(
		endpointSet.GetSourceEndpoint,
		decodeGetSourceRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	deleteSourceHandler := kithttp.NewServer(
		endpointSet.DeleteSourceEndpoint,
		decodeDeleteSourceRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	updateSourceHandler := kithttp.NewServer(
		endpointSet.UpdateSourceAsyncRequestEndpoint,
		decodeUpdateSourceAsyncRequest,
		encodeResponse(http.StatusAccepted),
		opts...,
	)

	getUpdateSourceStatus := kithttp.NewServer(
		endpointSet.GetUpdateSourceStatusEndpoint,
		decodeGetUpdateSourceStatusRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	// Connectors

	listConnectorsHandler := kithttp.NewServer(
		endpointSet.ListConnectorsEndpoint,
		decodeListConnectorsRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	getConnectorHandler := kithttp.NewServer(
		endpointSet.GetConnectorEndpoint,
		decodeGetConnectorRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	updateConnectorHandler := kithttp.NewServer(
		endpointSet.UpdateConnectorEndpoint,
		decodeUpdateConnectorRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	deleteConnectorHandler := kithttp.NewServer(
		endpointSet.DeleteConnectorEndpoint,
		decodeDeleteConnectorRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	// Jobs

	listConnectorJobsHandler := kithttp.NewServer(
		endpointSet.ListConnectorJobsEndpoint,
		decodeListConnectorJobsRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	getConnectorJobHandler := kithttp.NewServer(
		endpointSet.GetConnectorJobEndpoint,
		decodeGetConnectorJobRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	getConnectorJobLogsHandler := kithttp.NewServer(
		endpointSet.GetConnectorJobLogsEndpoint,
		decodeGetConnectorJobRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	triggerJobHandler := kithttp.NewServer(
		endpointSet.TriggerConnectorJobEndpoint,
		decodeTriggerJobRequest,
		encodeResponse(http.StatusCreated),
		opts...,
	)

	webhookHandler := kithttp.NewServer(
		endpointSet.WebhookEndpoint,
		decodeWebhookRequest,
		encodeResponse(http.StatusOK),
		opts...,
	)

	r := chi.NewRouter()

	r.Route("/workspaces", func(r chi.Router) {
		r.Method(http.MethodGet, "/", listWorkspacesHandler)
		r.Method(http.MethodPost, "/", createWorkspaceHandler)

		r.Route("/{workspaceId}", func(r chi.Router) {
			r.Method(http.MethodGet, "/", getWorkspaceHandler)

			r.Route("/sources_definitions", func(r chi.Router) {
				r.Method(http.MethodGet, "/", listSourcesDefinitionsHandler)
				r.Method(http.MethodGet, "/{sourceDefinitionId}/configuration", getSourceDefinitionConfigurationHandler)
				r.Method(http.MethodPost, "/{sourceDefinitionId}/sources", createSourceFromDefinitionHandler)
				r.Method(http.MethodGet, "/{sourceDefinitionId}/sources/{ticketId}/status", getCreateSourceFromDefinitionStatus)
			})

			r.Route("/sources", func(r chi.Router) {
				r.Method(http.MethodPost, "/{sourceId}/connectors", createConnectorFromSourceHandler)
				r.Method(http.MethodGet, "/", listSourcesHandler)
				r.Method(http.MethodGet, "/{sourceId}", getSourceHandler)
				r.Method(http.MethodPut, "/{sourceId}", updateSourceHandler)
				r.Method(http.MethodGet, "/{sourceId}/status/{ticketId}", getUpdateSourceStatus)
				r.Method(http.MethodDelete, "/{sourceId}", deleteSourceHandler)
				r.Method(http.MethodGet, "/{sourceId}/tables", getSourceTablesHandler)
				r.Method(http.MethodGet, "/{sourceId}/tables/{ticketId}/status", getGetSourceTablesStatus)
			})

			r.Route("/connectors", func(r chi.Router) {
				r.Method(http.MethodGet, "/", listConnectorsHandler)
				r.Route("/{connectorId}", func(r chi.Router) {
					r.Method(http.MethodGet, "/", getConnectorHandler)
					r.Method(http.MethodPut, "/", updateConnectorHandler)
					r.Method(http.MethodDelete, "/", deleteConnectorHandler)
					r.Route("/jobs", func(r chi.Router) {
						r.Method(http.MethodGet, "/", listConnectorJobsHandler)
						r.Method(http.MethodPost, "/", triggerJobHandler)
						r.Method(http.MethodGet, "/{jobId}", getConnectorJobHandler)
						r.Method(http.MethodGet, "/{jobId}/logs", getConnectorJobLogsHandler)
					})
				})
			})
		})

	})

	r.Method(http.MethodPost, "/airbyte_internal/sync_webhook", webhookHandler)

	return r
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	statusCode := http.StatusInternalServerError
	if t, ok := err.(*dataextractor.ApiError); ok && http.StatusText(t.Code) != "" {
		statusCode = t.Code
	}

	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func decodeCreateWorkspaceRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body createWorkspaceRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body, nil
}

func decodeListWorkspacesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := listWorkspacesRequest{
		NameContains: "",
		Limit:        20,
		Offset:       0,
	}

	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")
	nameContains := r.URL.Query().Get("nameContains")

	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return nil, err
		}
		req.Limit = l
	}
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return nil, err
		}
		req.Offset = o
	}
	if nameContains != "" {
		req.NameContains = nameContains
	}
	return req, nil
}

func decodeGetWorkspaceRequest(_ context.Context, r *http.Request) (interface{}, error) {
	workspaceId, err := uuid.Parse(chi.URLParam(r, "workspaceId"))
	if err != nil {
		return nil, err
	}

	return getWorkspaceRequest{
		workspaceId: workspaceId,
	}, nil
}

func decodeListSourcesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	workspaceId, err := uuid.Parse(chi.URLParam(r, "workspaceId"))
	if err != nil {
		return nil, err
	}

	return listSourcesRequest{
		workspaceId: workspaceId,
	}, nil
}

func decodeGetSourceRequest(_ context.Context, r *http.Request) (interface{}, error) {
	sourceId, err := uuid.Parse(chi.URLParam(r, "sourceId"))
	if err != nil {
		return nil, err
	}

	return getSourceRequest{
		sourceId: sourceId,
	}, nil
}

func decodeDeleteSourceRequest(_ context.Context, r *http.Request) (interface{}, error) {
	sourceId, err := uuid.Parse(chi.URLParam(r, "sourceId"))
	if err != nil {
		return nil, err
	}

	return deleteSourceRequest{
		sourceId: sourceId,
	}, nil
}

func decodeUpdateSourceAsyncRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body updateSourceAsyncRequest

	sourceId, err := uuid.Parse(chi.URLParam(r, "sourceId"))
	if err != nil {
		return nil, err
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	body.SourceId = sourceId
	return body, nil
}

func decodeGetUpdateSourceStatusRequest(_ context.Context, r *http.Request) (interface{}, error) {
	ticketId, err := uuid.Parse(chi.URLParam(r, "ticketId"))
	if err != nil {
		return nil, err
	}

	return &getStatusRequest{
		ticketId: &ticketId,
	}, nil
}

func decodeListSourcesDefinitionsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	workspaceId, err := uuid.Parse(chi.URLParam(r, "workspaceId"))
	if err != nil {
		return nil, err
	}
	return listSourcesDefinitionsRequest{
		workspaceId: workspaceId,
	}, nil
}

func decodeCreateConnectorFromSourceRequest(_ context.Context, r *http.Request) (interface{}, error) {
	workspaceId, err := uuid.Parse(chi.URLParam(r, "workspaceId"))
	if err != nil {
		return nil, err
	}
	sourceId, err := uuid.Parse(chi.URLParam(r, "sourceId"))
	if err != nil {
		return nil, err
	}
	var body createConnectorFromSourceRequestBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	body.SourceId = &sourceId

	return createConnectorFromSourceRequest{
		connector:   body,
		WorkspaceId: workspaceId,
	}, nil
}

func decodeGetCreateConnectorRequest(_ context.Context, r *http.Request) (interface{}, error) {
	ticketId, err := uuid.Parse(chi.URLParam(r, "ticketId"))
	if err != nil {
		return nil, err
	}
	return getCreateConnectorResponsRequest{
		ticketId: &ticketId,
	}, nil
}

func decodeUpdateConnectorRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body updateConnectorRequest

	connectorId, err := uuid.Parse(chi.URLParam(r, "connectorId"))
	if err != nil {
		return nil, err
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	body.ID = &connectorId
	return body, nil
}

func decodeGetConnectorRequest(_ context.Context, r *http.Request) (interface{}, error) {
	connectorId, err := uuid.Parse(chi.URLParam(r, "connectorId"))
	if err != nil {
		return nil, err
	}

	return getConnectorRequest{
		connectionId: connectorId,
	}, nil
}

func decodeCreateSourceFromDefinitionAsyncRequest(_ context.Context, r *http.Request) (interface{}, error) {
	workspaceId, err := uuid.Parse(chi.URLParam(r, "workspaceId"))
	if err != nil {
		return nil, err
	}

	var body createSourceFromDefinitionAsyncRequestBody

	sourceDefinitionId, err := uuid.Parse(chi.URLParam(r, "sourceDefinitionId"))
	if err != nil {
		return nil, err
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	body.SourceDefinitionId = sourceDefinitionId
	return createSourceFromDefinitionAsyncRequest{
		Source:      body,
		WorkspaceId: workspaceId,
	}, nil
}

func decodeGetStatusRequest(_ context.Context, r *http.Request) (interface{}, error) {
	ticketId, err := uuid.Parse(chi.URLParam(r, "ticketId"))
	if err != nil {
		return nil, err
	}

	return &getStatusRequest{
		ticketId: &ticketId,
	}, nil
}

func decodeGetSourceTablesAsyncRequest(_ context.Context, r *http.Request) (interface{}, error) {
	sourceId, err := uuid.Parse(chi.URLParam(r, "sourceId"))
	if err != nil {
		return nil, err
	}

	return getSourceTablesAsyncRequest(&sourceId), nil
}

func decodeListConnectorsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	workspaceId, err := uuid.Parse(chi.URLParam(r, "workspaceId"))
	if err != nil {
		return nil, err
	}

	req := listConnectorsRequest{
		WorkspaceId: workspaceId,
		Limit:       20,
		Offset:      0,
	}

	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return nil, err
		}
		req.Limit = l
	}
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return nil, err
		}
		req.Offset = o
	}

	return req, nil
}

func decodeDeleteConnectorRequest(_ context.Context, r *http.Request) (interface{}, error) {
	connectorId, err := uuid.Parse(chi.URLParam(r, "connectorId"))
	if err != nil {
		return nil, err
	}

	return deleteConnectorRequest{
		connectionId: connectorId,
	}, nil
}

func decodeGetSourceDefinitionConfigurationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	workspaceId, err := uuid.Parse(chi.URLParam(r, "workspaceId"))
	if err != nil {
		return nil, err
	}

	SourceDefinitionId := chi.URLParam(r, "sourceDefinitionId")
	return getSourceDefinitionSpecificationRequest{
		SourceDefinitionId: SourceDefinitionId,
		workspaceId:        workspaceId,
	}, nil
}

func decodeListConnectorJobsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := listJobsRequest{
		Limit:  20,
		Offset: 0,
	}
	connectorId, err := uuid.Parse(chi.URLParam(r, "connectorId"))
	if err != nil {
		return nil, err
	}

	req.ConnectorId = &connectorId

	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")
	jobType := r.URL.Query().Get("jobType")
	status := r.URL.Query().Get("status")
	createdAtStart := r.URL.Query().Get("createdAtStart")
	createdAtEnd := r.URL.Query().Get("createdAtEnd")
	updatedAtStart := r.URL.Query().Get("updatedAtStart")
	updatedAtEnd := r.URL.Query().Get("updatedAtEnd")
	orderBy := r.URL.Query().Get("orderBy")

	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return nil, err
		}
		req.Limit = l
	}

	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return nil, err
		}
		req.Offset = o
	}

	req.JobType = jobType
	req.Status = status
	req.CreatedAtStart = createdAtStart
	req.CreatedAtEnd = createdAtEnd
	req.OrderBy = orderBy
	req.UpdatedAtStart = updatedAtStart
	req.UpdatedAtEnd = updatedAtEnd
	return req, nil
}

func decodeGetConnectorJobRequest(_ context.Context, r *http.Request) (interface{}, error) {
	connectorId, err := uuid.Parse(chi.URLParam(r, "connectorId"))
	if err != nil {
		return nil, err
	}

	jobId, err := strconv.ParseInt(chi.URLParam(r, "jobId"), 10, 64)
	if err != nil {
		return nil, err
	}

	return getConnectorJobRequest{
		ConnectorId: &connectorId,
		JobId:       jobId,
	}, nil
}

func decodeTriggerJobRequest(_ context.Context, r *http.Request) (interface{}, error) {
	connectorId, err := uuid.Parse(chi.URLParam(r, "connectorId"))
	if err != nil {
		return nil, err
	}

	return triggerJobRequest{
		ConnectionId: connectorId,
	}, nil
}

func decodeWebhookRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req webhookRequest
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	err := json.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func encodeResponse(statusCode int) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, response interface{}) error {
		if e, ok := response.(errorer); ok && e.error() != nil {
			encodeError(ctx, e.error(), w)
			return nil
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if http.StatusText(statusCode) != "" {
			w.WriteHeader(statusCode)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		return json.NewEncoder(w).Encode(response)
	}
}

type errorer interface {
	error() error
}
