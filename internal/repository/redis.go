package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	_, err := client.Ping(ctx).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	slog.Info("Connected to Redis", "addr", addr)

	return client, nil
}
