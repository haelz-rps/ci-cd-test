package repository

import "context"

type Repository interface {
	Dial(ctx context.Context, host string) error
	Close(ctx context.Context) error
	Migrate(ctx context.Context, models ...interface{}) error
}
