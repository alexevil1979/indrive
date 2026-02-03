// Package redis — Redis client for rate limiting, sessions (2026)
// Optional: if REDIS_ADDR empty, service runs without Redis (e.g. in-memory rate limit)
package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Addr empty = no Redis (optional for local dev)

// New creates a Redis client. Addr empty = no Redis (returns nil; optional for local dev).
func New(addr string) (*redis.Client, error) {
	if addr == "" {
		return nil, nil
	}
	cli := redis.NewClient(&redis.Options{
		Addr:         addr,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	if err := cli.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return cli, nil
}
