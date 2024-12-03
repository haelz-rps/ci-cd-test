package redis

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/ubix/dripper/manager/repositories/kvstore"
)

type redisStore struct {
	client *redis.Client
}

func NewRedisStore(redisUrl string) (*redisStore, error) {
	opts, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	return &redisStore{client}, nil
}

func (i *redisStore) Get(ctx context.Context, key uuid.UUID) (interface{}, error) {
	bytesValue, err := i.client.Get(ctx, key.String()).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, kvstore.ErrKeyNotFound
		} else {
			return nil, err
		}
	}
	var value interface{}
	err = json.Unmarshal(bytesValue, &value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (i *redisStore) Set(ctx context.Context, key uuid.UUID, value interface{}) error {
	stringValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = i.client.Set(ctx, key.String(), stringValue, 0).Err()
	return err
}

func (i *redisStore) Close(ctx context.Context) error {
	return nil
}
