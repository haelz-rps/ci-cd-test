package inmem

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/ubix/dripper/manager/repositories/kvstore"
)

type inmemStore struct {
	m     sync.RWMutex
	store map[uuid.UUID]interface{}
}

func NewInmemStore() *inmemStore {
	return &inmemStore{
		store: make(map[uuid.UUID]interface{}),
	}
}

func (i *inmemStore) Get(ctx context.Context, key uuid.UUID) (interface{}, error) {
	i.m.Lock()
	defer i.m.Unlock()

	value, exists := i.store[key]
	if !exists {
		return nil, kvstore.ErrKeyNotFound
	}
	return value, nil
}

func (i *inmemStore) Set(ctx context.Context, key uuid.UUID, value interface{}) error {
	i.m.Lock()
	defer i.m.Unlock()

	i.store[key] = value

	return nil
}

func (i *inmemStore) Close(ctx context.Context) error {
	return nil
}
