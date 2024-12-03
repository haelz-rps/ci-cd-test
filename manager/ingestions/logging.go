package ingestions

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/uuid"

	dataextractor "github.com/ubix/dripper/manager/repositories/data-extractor"
)

type loggingMiddleware struct {
	logger log.Logger
	cs     ConnectorService
}

func NewLoggingMiddleware(logger log.Logger, s ConnectorService) ConnectorService {
	return &loggingMiddleware{
		logger: logger,
		cs:     s,
	}
}

func (s *loggingMiddleware) UpdateConnector(ctx context.Context, c *dataextractor.Connector) (*dataextractor.Connector, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"connector", c.ID,
			"method", "UpdateConnector",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.UpdateConnector(ctx, c)
}

func (s *loggingMiddleware) ListWorkspaces(ctx context.Context, pag dataextractor.Filters) ([]*dataextractor.Workspace, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"method", "ListWorkspaces",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.ListWorkspaces(ctx, pag)
}

func (s *loggingMiddleware) GetWorkspace(ctx context.Context, workspaceId uuid.UUID) (*dataextractor.Workspace, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"workspaceId", workspaceId,
			"method", "GetWorkspace",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.GetWorkspace(ctx, workspaceId)
}

func (s *loggingMiddleware) CreateWorkspace(ctx context.Context, tenant, bucket string) (*dataextractor.Workspace, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"tenant", tenant,
			"method", "CreateWorkspace",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.CreateWorkspace(ctx, tenant, bucket)
}

func (s *loggingMiddleware) CreateConnectorFromSource(ctx context.Context, workspaceId uuid.UUID, c *dataextractor.Connector) (*dataextractor.Connector, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"workspaceId", workspaceId,
			"source", c.SourceId,
			"method", "CreateConnectorFromSource",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.CreateConnectorFromSource(ctx, workspaceId, c)
}

func (s *loggingMiddleware) GetConnector(ctx context.Context, connectorId uuid.UUID) (*dataextractor.Connector, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"connector", connectorId,
			"method", "GetConnector",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.GetConnector(ctx, connectorId)
}

func (s *loggingMiddleware) ListConnectors(ctx context.Context, filter *dataextractor.Filters) ([]*dataextractor.Connector, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"method", "ListConnectors",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.ListConnectors(ctx, filter)
}

func (s *loggingMiddleware) DeleteConnector(ctx context.Context, connectorId uuid.UUID) error {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"connector", connectorId,
			"method", "DeleteConnector",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.DeleteConnector(ctx, connectorId)
}

func (s *loggingMiddleware) ListSourcesDefinitions(ctx context.Context, workspaceId uuid.UUID) ([]*dataextractor.SourceDefinition, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"method", "ListSourcesDefinitions",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.ListSourcesDefinitions(ctx, workspaceId)
}

func (s *loggingMiddleware) GetSourceDefinitionConfiguration(ctx context.Context, workspaceId uuid.UUID, sourceDefId string) (*dataextractor.SourceDefinitionConfiguration, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"sourceDefinitionId", sourceDefId,
			"method", "GetSourceDefinitionConfiguration",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.GetSourceDefinitionConfiguration(ctx, workspaceId, sourceDefId)
}

func (s *loggingMiddleware) CreateSourceFromDefinitionAsync(ctx context.Context, workspaceId uuid.UUID, source *dataextractor.Source) (*uuid.UUID, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"method", "CreateSourceFromDefinitionAsync",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.CreateSourceFromDefinitionAsync(ctx, workspaceId, source)
}

func (s *loggingMiddleware) CreateSourceFromDefinition(ctx context.Context, workspaceId uuid.UUID, source *dataextractor.Source, ticketId *uuid.UUID) error {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"source", source.SourceId,
			"ticket", ticketId,
			"method", "CreateSourceFromDefinition",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.CreateSourceFromDefinition(ctx, workspaceId, source, ticketId)
}

func (s *loggingMiddleware) GetCreateSourceFromDefinitionStatus(ctx context.Context, ticketId *uuid.UUID) (*CreateSourceStatus, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"ticket", ticketId,
			"method", "GetCreateSourceFromDefinitionStatus",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.GetCreateSourceFromDefinitionStatus(ctx, ticketId)
}

func (s *loggingMiddleware) GetSource(ctx context.Context, sourceId uuid.UUID) (*dataextractor.Source, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"source", sourceId,
			"method", "GetSource",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.GetSource(ctx, sourceId)
}

func (s *loggingMiddleware) UpdateSourceAsync(ctx context.Context, source *dataextractor.Source) (*uuid.UUID, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"source", source.SourceId,
			"method", "UpdateSourceAsync",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.UpdateSourceAsync(ctx, source)
}

func (s *loggingMiddleware) UpdateSource(ctx context.Context, source *dataextractor.Source, ticketId *uuid.UUID) (*dataextractor.Source, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"source", source.SourceId,
			"ticket", ticketId,
			"method", "UpdateSource",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.UpdateSource(ctx, source, ticketId)
}

func (s *loggingMiddleware) GetUpdateSourceStatus(ctx context.Context, ticketId *uuid.UUID) (*UpdateSourceStatus, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"ticket", ticketId,
			"method", "GetUpdateSourceStatus",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.GetUpdateSourceStatus(ctx, ticketId)
}

func (s *loggingMiddleware) DeleteSource(ctx context.Context, sourceId uuid.UUID) error {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"source", sourceId,
			"method", "DeleteSources",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.DeleteSource(ctx, sourceId)
}

func (s *loggingMiddleware) GetSourceTables(ctx context.Context, sourceId *uuid.UUID, ticketId *uuid.UUID) ([]*dataextractor.Table, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"source", sourceId,
			"ticket", ticketId,
			"method", "GetSourceTables",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.GetSourceTables(ctx, sourceId, ticketId)
}

func (s *loggingMiddleware) GetGetSourceTablesStatus(ctx context.Context, ticketId *uuid.UUID) (*GetSourceTablesStatus, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"ticket", ticketId,
			"method", "GetGetSourceTablesStatus",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.GetGetSourceTablesStatus(ctx, ticketId)
}

func (s *loggingMiddleware) GetSourceTablesAsync(ctx context.Context, sourceId *uuid.UUID) (*uuid.UUID, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"source", sourceId,
			"method", "GetSourceTablesAsync",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.GetSourceTablesAsync(ctx, sourceId)
}

func (s *loggingMiddleware) ListSources(ctx context.Context, workspaceId uuid.UUID) ([]*dataextractor.Source, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"method", "ListSources",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.ListSources(ctx, workspaceId)
}

func (s *loggingMiddleware) ListConnectorJobs(ctx context.Context, connectorId *uuid.UUID, filter *dataextractor.Filters) ([]*dataextractor.Job, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"method", "ListConnectorJobs",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.ListConnectorJobs(ctx, connectorId, filter)
}

func (s *loggingMiddleware) GetConnectorJob(ctx context.Context, jobId int64) (*dataextractor.Job, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"method", "GetConnectorJob",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.GetConnectorJob(ctx, jobId)
}

func (s *loggingMiddleware) GetConnectorJobLogs(ctx context.Context, jobId int64) ([]string, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"method", "GetConnectorJobLogs",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.GetConnectorJobLogs(ctx, jobId)
}

func (s *loggingMiddleware) TriggerJob(ctx context.Context, connectorId uuid.UUID) (*dataextractor.Job, error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"connector", connectorId,
			"method", "TriggerJob",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.TriggerJob(ctx, connectorId)
}

func (s *loggingMiddleware) SendWebhookToKafka(ctx context.Context, connectorId uuid.UUID, workspaceId uuid.UUID) error {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"workspace", workspaceId,
			"connector", connectorId,
			"method", "SendWebhookToKafka",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.SendWebhookToKafka(ctx, connectorId, workspaceId)
}

func (s *loggingMiddleware) TriggerProfiling(ctx context.Context, connectorId uuid.UUID, workspaceId uuid.UUID) error {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"workspace", workspaceId,
			"connector", connectorId,
			"method", "TriggerProfiling",
			"took", time.Since(begin),
		)
	}(time.Now())
	return s.cs.TriggerProfiling(ctx, connectorId, workspaceId)
}
