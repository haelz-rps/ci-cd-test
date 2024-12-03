package datacalog

import (
	"context"

	dataextractor "github.com/ubix/dripper/manager/repositories/data-extractor"
)

type DataCatalog interface {
	CreateDatabase(context.Context, *dataextractor.Connector) error
	ProfileTables(context.Context, *dataextractor.Connector) error
	DeleteDatabase(context.Context, *dataextractor.Connector) error
	RefreshTablesPartitions(context.Context, *dataextractor.Connector, string) error
	UpdateTables(context.Context, *dataextractor.Connector) error
}
