// Package redis implements the store.Store interface using Redis.
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hauk/hauk-go/internal/config"
	"github.com/hauk/hauk-go/internal/store"
	"github.com/redis/go-redis/v9"
)

// Store implements the store.Store interface using Redis.
type Store struct {
	client *redis.Client
	prefix string
}

// NewStore creates a new Redis store with the given config.
func NewStore(ctx context.Context, cfg *config.Config) (*Store, error) {
	var addr string
	if cfg.RedisHost[0] == '/' {
		// UNIX socket
		addr = cfg.RedisHost
	} else {
		addr = fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.RedisAuth,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Store{
		client: rdb,
		prefix: cfg.RedisPrefix,
	}, nil
}

func (s *Store) key(key string) string {
	return s.prefix + key
}

// Get retrieves a value by key.
func (s *Store) Get(key string) (interface{}, bool) {
	ctx := context.Background()
	fullKey := s.key(key)

	val, err := s.client.Get(ctx, fullKey).Result()
	if err == redis.Nil {
		return nil, false
	}
	if err != nil {
		return nil, false
	}

	var data interface{}
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, false
	}

	return data, true
}

// Set stores a value with the given key, expiring after expire seconds.
func (s *Store) Set(key string, data interface{}, expire int) error {
	ctx := context.Background()
	fullKey := s.key(key)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if expire > 0 {
		return s.client.SetEx(ctx, fullKey, jsonData, time.Duration(expire)*time.Second).Err()
	}
	return s.client.Set(ctx, fullKey, jsonData, 0).Err()
}

// Delete removes the value at the given key.
func (s *Store) Delete(key string) error {
	ctx := context.Background()
	fullKey := s.key(key)
	return s.client.Del(ctx, fullKey).Err()
}

// Close closes the Redis connection.
func (s *Store) Close() error {
	return s.client.Close()
}

var _ store.Store = (*Store)(nil)
