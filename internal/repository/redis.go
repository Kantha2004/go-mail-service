package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func NewRediusClient(addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	_, err := client.Ping(Ctx).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	slog.Info("Connected to Redis", "addr", addr)

	return client, nil
}
