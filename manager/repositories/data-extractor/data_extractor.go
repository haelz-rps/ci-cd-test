package dataextractor

import (
	"context"

	"github.com/google/uuid"
)

type Filters struct {
	Limit          int
	Offset         int
	ConnectionId   uuid.UUID
	JobType        string
	WorkspaceId    uuid.UUID
	Status         string
	CreatedAtStart string
	CreatedAtEnd   string
	UpdatedAtStart string
	UpdatedAtEnd   string
	OrderBy        string
	NameContains   string
}

type DataExtractor interface {
	ListWorkspaces(ctx context.Context, filter Filters) ([]*Workspace, error)
	GetWorkspace(ctx context.Context, workspaceId uuid.UUID) (*Workspace, error)
	CreateWorkspace(ctx context.Context, tenant, bucket string) (*Workspace, error)

	ListSourcesDefinitions(ctx context.Context, workspaceId uuid.UUID) ([]*SourceDefinition, error)
	GetSourceDefinitionConfiguration(ctx context.Context, workspaceId uuid.UUID, sourceDefinitionId string) (*SourceDefinitionConfiguration, error)
	CreateSourceFromDefinition(ctx context.Context, workspaceId uuid.UUID, s *Source) error

	ListSources(ctx context.Context, workspaceId uuid.UUID) ([]*Source, error)
	GetSource(ctx context.Context, sourceId uuid.UUID) (*Source, error)

	GetSourceTables(ctx context.Context, sourceId uuid.UUID) ([]*Table, error)
	UpdateSource(ctx context.Context, source *Source) (*Source, error)
	DeleteSource(ctx context.Context, sourceId uuid.UUID) error
	CreateConnectorFromSource(ctx context.Context, workspaceId uuid.UUID, connector *Connector) (*Connector, error)

	ListConnectors(context.Context, *Filters) ([]*Connector, error)
	GetConnector(ctx context.Context, connectorId uuid.UUID) (*Connector, error)
	UpdateConnector(context.Context, *Connector) (*Connector, error)
	DeleteConnector(ctx context.Context, connectorId uuid.UUID) error

	TriggerConnectorJob(ctx context.Context, connectionId uuid.UUID) (*Job, error)
	ListConnectorJobs(ctx context.Context, connectorId *uuid.UUID, filters *Filters) ([]*Job, error)
	GetConnectorJobDetails(ctx context.Context, jobId int64) (*Job, error)
	GetConnectorJobLogs(ctx context.Context, jobId int64) ([]string, error)
}
