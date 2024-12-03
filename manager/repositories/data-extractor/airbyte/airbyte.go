package airbyte

import (
	"context"
	"fmt"
	"time"

	"net/http"

	"github.com/gojek/heimdall/v7/httpclient"

	"github.com/google/uuid"
	"github.com/ubix/dripper/manager/configs"
	dataextractor "github.com/ubix/dripper/manager/repositories/data-extractor"
	airbyteclient "github.com/ubix/dripper/manager/repositories/data-extractor/airbyte/airbyte-client"
)

type airbyte struct {
	client *airbyteclient.ClientWithResponses
	config configs.AirbyteConfig
}

func NewAirbyte(config configs.AirbyteConfig) (dataextractor.DataExtractor, error) {
	httpClient := httpclient.NewClient(
		httpclient.WithHTTPTimeout(30 * time.Minute),
	)
	customHttpClient := airbyteclient.WithHTTPClient(httpClient)
	client, err := airbyteclient.NewClientWithResponses(config.Host, customHttpClient)
	if err != nil {
		return nil, err
	}
	return &airbyte{client, config}, nil
}

func (a *airbyte) getDestinationId(ctx context.Context, workspaceId uuid.UUID) (*uuid.UUID, error) {
	res, err := a.client.ListDestinationsForWorkspaceWithResponse(
		ctx,
		airbyteclient.WorkspaceIdRequestBody{WorkspaceId: workspaceId},
	)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	if len(res.JSON200.Destinations) > 1 {
		return nil, fmt.Errorf("Airbyte should have only 1 destination. Instead it has: %d", len(res.JSON200.Destinations))
	}

	return &res.JSON200.Destinations[0].DestinationId, nil
}

func (a *airbyte) checkConnection(ctx context.Context, workspaceId uuid.UUID, s *dataextractor.Source) error {
	var res *airbyteclient.ExecuteSourceCheckConnectionResponse
	var err error

	// Retry the endpoint to make sure it won't fail due to pod startup delay
	shouldTryAgain := true
	retries := 3
	for i := 0; i < retries && shouldTryAgain; i++ {
		// Add a sleep time between retries
		time.Sleep(time.Duration(i) * 100 * time.Millisecond)

		res, err = a.client.ExecuteSourceCheckConnectionWithResponse(ctx, airbyteclient.SourceCoreConfig{
			ConnectionConfiguration: *s.Configuration,
			SourceDefinitionId:      s.SourceDefinitionId,
			WorkspaceId:             &workspaceId,
		})

		if err == nil && res != nil && res.StatusCode() == http.StatusOK {
			shouldTryAgain = false
		}
	}

	if err != nil {
		return err
	}

	if res.StatusCode() == http.StatusUnprocessableEntity {
		return dataextractor.NewApiError(res.StatusCode(), []byte(res.JSON422.Message))
	} else if res.StatusCode() != http.StatusOK {
		return dataextractor.NewApiError(res.StatusCode(), res.Body)
	} else if res.JSON200.Status != airbyteclient.CheckConnectionReadStatusSucceeded {
		return dataextractor.NewApiError(res.StatusCode(), []byte(*res.JSON200.Message))
	}

	return nil
}

type AirbyteConnectionError struct {
	message string
}

func (a *airbyte) ListWorkspaces(ctx context.Context, filter dataextractor.Filters) ([]*dataextractor.Workspace, error) {
	body := airbyteclient.ListResourcesForWorkspacesRequestBody{
		Pagination: airbyteclient.Pagination{
			PageSize:  filter.Limit,
			RowOffset: filter.Offset,
		},
	}
	if filter.NameContains != "" {
		body.NameContains = &filter.NameContains
	}
	res, err := a.client.ListAllWorkspacesPaginatedWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	workspaces := make([]*dataextractor.Workspace, len(res.JSON200.Workspaces))
	for i, ws := range res.JSON200.Workspaces {
		workspaces[i] = &dataextractor.Workspace{
			ID:   &ws.WorkspaceId,
			Name: ws.Name,
		}
	}
	return workspaces, nil
}

func (a *airbyte) GetWorkspace(ctx context.Context, workspaceId uuid.UUID) (*dataextractor.Workspace, error) {
	res, err := a.client.GetWorkspaceWithResponse(
		ctx,
		airbyteclient.GetWorkspaceJSONRequestBody{WorkspaceId: workspaceId},
	)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	return &dataextractor.Workspace{
		ID:   &res.JSON200.WorkspaceId,
		Name: res.JSON200.Name,
	}, nil
}

func (a *airbyte) setupDestination(ctx context.Context, workspaceId uuid.UUID, bucketName string) error {
	definitionId, err := uuid.Parse(a.config.DestinationTemplates.S3.DefinitionId)
	if err != nil {
		return err
	}

	updateRes, err := a.client.UpdateDestinationDefinitionWithResponse(
		ctx,
		airbyteclient.UpdateDestinationDefinitionJSONRequestBody{
			DestinationDefinitionId: definitionId,
			DockerImageTag:          a.config.DestinationTemplates.S3.DockerImageTag,
		},
	)
	if err != nil {
		return err
	}
	if updateRes.StatusCode() != http.StatusOK {
		return dataextractor.NewApiError(updateRes.StatusCode(), updateRes.Body)
	}

	configuration := a.config.DestinationTemplates.S3.Config
	configuration[a.config.DestinationTemplates.S3.BucketKey] = bucketName

	res, err := a.client.CreateDestinationWithResponse(ctx, airbyteclient.DestinationCreate{
		Name:                    bucketName,
		WorkspaceId:             workspaceId,
		DestinationDefinitionId: definitionId,
		ConnectionConfiguration: configuration,
	})
	if err != nil {
		return err
	}

	if res.StatusCode() != http.StatusOK {
		return dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	return nil
}

func (a *airbyte) setupNotificationWebhook(ctx context.Context, workspaceId uuid.UUID) error {
	res, err := a.client.UpdateWorkspaceWithResponse(ctx, airbyteclient.WorkspaceUpdate{
		WorkspaceId: workspaceId,
		NotificationSettings: &airbyteclient.NotificationSettings{
			SendOnSuccess: airbyteclient.NotificationSettingsSendOn{
				NotificationType: &[]airbyteclient.NotificationType{"slack"},
				SlackConfiguration: &airbyteclient.SlackNotificationConfiguration{
					Webhook: a.config.SyncWebhook,
				},
			},
		},
	})
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	return nil
}

func (a *airbyte) CreateWorkspace(ctx context.Context, tenant, bucket string) (*dataextractor.Workspace, error) {
	organizationId, err := uuid.Parse(a.config.OrganizationId)
	if err != nil {
		return nil, err
	}

	res, err := a.client.CreateWorkspaceWithResponse(
		ctx,
		airbyteclient.WorkspaceCreate{
			Name:           tenant,
			OrganizationId: organizationId,
		},
	)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	workspaceId := res.JSON200.WorkspaceId

	err = a.setupDestination(ctx, workspaceId, bucket)
	if err != nil {
		return nil, err
	}

	err = a.setupNotificationWebhook(ctx, workspaceId)
	if err != nil {
		return nil, err
	}

	return &dataextractor.Workspace{
		ID:   &res.JSON200.WorkspaceId,
		Name: res.JSON200.Name,
	}, nil
}

func (a *airbyte) ListSourcesDefinitions(ctx context.Context, workspaceId uuid.UUID) ([]*dataextractor.SourceDefinition, error) {
	res, err := a.client.ListSourceDefinitionsForWorkspaceWithResponse(ctx, airbyteclient.WorkspaceIdRequestBody{
		WorkspaceId: workspaceId,
	})
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	sds := make([]*dataextractor.SourceDefinition, len(res.JSON200.SourceDefinitions))
	for i, sd := range res.JSON200.SourceDefinitions {
		sds[i] = &dataextractor.SourceDefinition{
			SourceDefinitionId: *sd.SourceDefinitionId,
			DocumentationUrl:   *sd.DocumentationUrl,
			Name:               *sd.Name,
		}
		if sd.Icon != nil {
			sds[i].Icon = *sd.Icon
		}
	}
	return sds, nil
}

func (a *airbyte) GetSourceDefinitionConfiguration(ctx context.Context, workspaceId uuid.UUID, id string) (*dataextractor.SourceDefinitionConfiguration, error) {
	sourceDefinitionId, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	res, err := a.client.GetSourceDefinitionSpecificationWithResponse(
		ctx,
		airbyteclient.SourceDefinitionIdWithWorkspaceIdRequestBody{
			SourceDefinitionId: sourceDefinitionId,
			WorkspaceId:        &workspaceId,
		})
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	return &dataextractor.SourceDefinitionConfiguration{
		SourceDefinitionId:  &res.JSON200.SourceDefinitionId,
		DocumentationUrl:    res.JSON200.DocumentationUrl,
		SourceConfiguration: (*dataextractor.SourceConfiguration)(res.JSON200.ConnectionSpecification),
	}, nil
}

func (a *airbyte) CreateSourceFromDefinition(ctx context.Context, workspaceId uuid.UUID, source *dataextractor.Source) error {
	err := a.checkConnection(ctx, workspaceId, source)
	if err != nil {
		return err
	}

	res, err := a.client.CreateSourceWithResponse(ctx, airbyteclient.SourceCreate{
		ConnectionConfiguration: *source.Configuration,
		SourceDefinitionId:      source.SourceDefinitionId,
		WorkspaceId:             workspaceId,
		Name:                    source.Name,
	})
	if err != nil {
		return err
	}

	if res.StatusCode() != http.StatusOK {
		return dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	source.SourceId = res.JSON200.SourceId

	return nil
}

func (a *airbyte) ListSources(ctx context.Context, workspaceId uuid.UUID) ([]*dataextractor.Source, error) {
	res, err := a.client.ListSourcesForWorkspaceWithResponse(ctx, airbyteclient.ListSourcesForWorkspaceJSONRequestBody{
		WorkspaceId: workspaceId,
	})
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}
	return ParseSourceReadList(res.JSON200), nil
}

func (a *airbyte) GetSource(ctx context.Context, sourceId uuid.UUID) (*dataextractor.Source, error) {
	res, err := a.client.GetSourceWithResponse(ctx, airbyteclient.GetSourceJSONRequestBody{
		SourceId: sourceId,
	})
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	source := ParseSourceRead(*res.JSON200)

	if source.Configuration == nil || len(*source.Configuration) == 0 {
		return nil, dataextractor.NewApiError(404, []byte("source not found"))
	}

	return source, nil
}

func (a *airbyte) GetSourceTables(ctx context.Context, sourceId uuid.UUID) ([]*dataextractor.Table, error) {
	var res *airbyteclient.DiscoverSchemaForSourceResponse
	var err error

	body := airbyteclient.SourceIdRequestBody{SourceId: sourceId}

	shouldTryAgain := true
	retries := 3
	for i := 0; i < retries && shouldTryAgain; i++ {
		res, err = a.client.DiscoverSchemaForSourceWithResponse(ctx, body)

		if err == nil && res != nil && res.StatusCode() == http.StatusOK {
			shouldTryAgain = false
		}
	}

	if err != nil {
		return nil, err
	}

	if res.StatusCode() == http.StatusUnprocessableEntity {
		return nil, dataextractor.NewApiError(res.StatusCode(), []byte(res.JSON422.Message))
	} else if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	tables := ParseSyncCatalogToTables(res.JSON200.Catalog)

	return tables, nil
}

func (a *airbyte) UpdateSource(ctx context.Context, source *dataextractor.Source) (*dataextractor.Source, error) {
	checkConnReq := airbyteclient.CheckConnectionToSourceForUpdateJSONRequestBody{
		Name:                    source.Name,
		SourceId:                source.SourceId,
		ConnectionConfiguration: *source.Configuration,
	}

	var checkConnRes *airbyteclient.CheckConnectionToSourceForUpdateResponse
	var err error

	shouldTryAgain := true
	retries := 3
	for i := 0; i < retries && shouldTryAgain; i++ {
		checkConnRes, err = a.client.CheckConnectionToSourceForUpdateWithResponse(ctx, checkConnReq)

		if err == nil || checkConnRes.StatusCode() != http.StatusOK {
			shouldTryAgain = false
		}
	}
	if err != nil {
		return nil, err
	}
	if checkConnRes.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(checkConnRes.StatusCode(), checkConnRes.Body)
	}

	res, err := a.client.UpdateSourceWithResponse(ctx, airbyteclient.UpdateSourceJSONRequestBody{
		SourceId:                source.SourceId,
		Name:                    source.Name,
		ConnectionConfiguration: *source.Configuration,
	})
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	return ParseSourceRead(*res.JSON200), nil
}

func (a *airbyte) DeleteSource(ctx context.Context, sourceId uuid.UUID) error {
	res, err := a.client.DeleteSourceWithResponse(ctx, airbyteclient.DeleteSourceJSONRequestBody{
		SourceId: sourceId,
	})
	if err != nil {
		return err
	}

	if res.StatusCode() != http.StatusNoContent {
		return dataextractor.NewApiError(res.StatusCode(), res.Body)
	}
	return nil
}

func (a *airbyte) CreateConnectorFromSource(ctx context.Context, workspaceId uuid.UUID, connector *dataextractor.Connector) (*dataextractor.Connector, error) {
	destinationId, err := a.getDestinationId(ctx, workspaceId)
	if err != nil {
		return nil, err
	}

	body := GenerateCreateConnectionRequestBody(connector, destinationId)

	res, err := a.client.WebBackendCreateConnectionWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	connector = ParseConnectionRead(res.JSON200)

	return connector, nil
}

func (a *airbyte) UpdateConnector(ctx context.Context, connector *dataextractor.Connector) (*dataextractor.Connector, error) {
	res, err := a.client.WebBackendUpdateConnectionWithResponse(
		ctx,
		GenerateUpdateConnectionRequestBody(connector),
	)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	connector = ParseConnectionRead(res.JSON200)

	return connector, nil
}

func (a *airbyte) ListConnectors(ctx context.Context, pagination *dataextractor.Filters) ([]*dataextractor.Connector, error) {
	res, err := a.client.ListConnectionsForWorkspaceWithResponse(ctx, airbyteclient.ListConnectionsForWorkspaceJSONRequestBody{
		WorkspaceId: pagination.WorkspaceId,
	})
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	cp := ParseListConnectionsForWorkspaceWithResponse(res.JSON200)
	return cp, nil
}

func (a *airbyte) GetConnector(ctx context.Context, connectorId uuid.UUID) (*dataextractor.Connector, error) {
	res, err := a.client.WebBackendGetConnectionWithResponse(ctx, airbyteclient.WebBackendConnectionRequestBody{
		ConnectionId: connectorId,
	})
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	conn := ParseConnectionRead(res.JSON200)
	return conn, nil
}

func (a *airbyte) DeleteConnector(ctx context.Context, connectorId uuid.UUID) error {
	res, err := a.client.DeleteConnectionWithResponse(ctx, airbyteclient.DeleteConnectionJSONRequestBody{
		ConnectionId: connectorId,
	})
	if err != nil {
		return err
	}

	if res.StatusCode() != http.StatusNoContent {
		return dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	return nil
}

func (a *airbyte) TriggerConnectorJob(ctx context.Context, connectorId uuid.UUID) (*dataextractor.Job, error) {
	req := airbyteclient.JobCreateRequest{
		ConnectionId: connectorId,
		JobType:      airbyteclient.JobTypeEnumSync,
	}
	res, err := a.client.PublicCreateJobWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	// TODO: use parser functions for this later
	return &dataextractor.Job{
		Id:     res.JSON200.JobId,
		Status: string(res.JSON200.Status),
	}, nil
}

func (a *airbyte) ListConnectorJobs(ctx context.Context, connectorId *uuid.UUID, filter *dataextractor.Filters) ([]*dataextractor.Job, error) {
	jobConfigTypes := []airbyteclient.JobConfigType{"sync", "reset_connection", "clear", "refresh"}
	reqBody := airbyteclient.JobListRequestBody{
		ConfigId:    connectorId.String(),
		ConfigTypes: jobConfigTypes,
	}

	res, err := a.client.ListJobsForWithResponse(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	jobs := make([]*dataextractor.Job, len(res.JSON200.Jobs))
	for i, jobWithAttempt := range res.JSON200.Jobs {
		jobs[i] = &dataextractor.Job{
			Id:          jobWithAttempt.Job.Id,
			Status:      string(jobWithAttempt.Job.Status),
			CreatedAt:   jobWithAttempt.Job.CreatedAt,
			UpdatedAt:   jobWithAttempt.Job.UpdatedAt,
			RowsSynced:  jobWithAttempt.Job.AggregatedStats.RecordsCommitted,
			BytesSynced: jobWithAttempt.Job.AggregatedStats.BytesCommitted,
		}
	}
	return jobs, nil
}

func (a *airbyte) GetConnectorJobDetails(ctx context.Context, jobId int64) (*dataextractor.Job, error) {
	reqBody := airbyteclient.JobIdRequestBody{Id: airbyteclient.JobId(jobId)}

	res, err := a.client.GetJobInfoWithResponse(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	return &dataextractor.Job{
		Id:        res.JSON200.Job.Id,
		Status:    string(res.JSON200.Job.Status),
		CreatedAt: res.JSON200.Job.CreatedAt,
		UpdatedAt: res.JSON200.Job.UpdatedAt,
		// TODO: the /v1/jobs/get doesn't returns the AggregatedStats
		// RowsSynced:  res.JSON200.Job.AggregatedStats.RecordsCommitted,
		// BytesSynced: res.JSON200.Job.AggregatedStats.BytesCommitted,
	}, nil
}

func (a *airbyte) GetConnectorJobLogs(ctx context.Context, jobId int64) ([]string, error) {
	reqBody := airbyteclient.JobIdRequestBody{Id: airbyteclient.JobId(jobId)}

	res, err := a.client.GetJobInfoWithResponse(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		return nil, dataextractor.NewApiError(res.StatusCode(), res.Body)
	}

	// First get the total number of lines for the log
	logsLen := 0
	for _, attempt := range res.JSON200.Attempts {
		logsLen += len(attempt.Logs.LogLines)
	}

	// Create log array with the exact space
	logs := make([]string, logsLen)
	// Set slice start to 0 to know from where to start the copy in case of multiple attempts
	sliceStart := 0
	for _, attempt := range res.JSON200.Attempts {
		copy(logs[sliceStart:sliceStart+len(attempt.Logs.LogLines)], attempt.Logs.LogLines)
		sliceStart += len(attempt.Logs.LogLines)
	}

	return logs, nil
}
