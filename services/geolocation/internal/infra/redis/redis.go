package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const GeoKeyDrivers = "drivers:location"

// New creates Redis client. Addr required for geolocation (GEO commands).
func New(addr string) (*redis.Client, error) {
	if addr == "" {
		addr = "localhost:6379"
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
