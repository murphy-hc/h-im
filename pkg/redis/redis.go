// Package redis provides a Redis client wrapper.
package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

// Config holds Redis connection parameters.
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewClient creates a new Redis client and verifies connectivity.
func NewClient(ctx context.Context, cfg Config) (*goredis.Client, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis: connect: %w", err)
	}

	return rdb, nil
}
