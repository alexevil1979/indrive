package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

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
