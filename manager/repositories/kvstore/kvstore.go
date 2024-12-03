package kvstore

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type KVStore interface {
	Get(ctx context.Context, key uuid.UUID) (interface{}, error)
	Set(ctx context.Context, key uuid.UUID, value interface{}) error
	Close(ctx context.Context) error
}

// Common Errors
var (
	ErrKeyNotFound = errors.New("key not found")
)
