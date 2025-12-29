package service

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
)

const (
	EMAIL_STREAM   = "go-email:microservice"
	EMAIL_GROUP    = "go-email:group"
	EMAIL_CONSUMER = "go-email:consumer"
)

type EmailWorker struct {
	redisClient *redis.Client
	Handler     func(context.Context, redis.XMessage) error
}

func NewEmailWorker(client *redis.Client) *EmailWorker {
	return &EmailWorker{
		redisClient: client,
		Handler: func(ctx context.Context, msg redis.XMessage) error {
			slog.Info("Processing email message", "message_id", msg.ID, "values", msg.Values)
			return nil
		},
	}
}

func (w *EmailWorker) Start(ctx context.Context) {
	slog.Info("EmailWorker starting...")

	w.ensureGroupExists(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Email Worker stopping...")
			return
		default:
			w.processNextBatch(ctx)
		}
	}
}

func (w *EmailWorker) ensureGroupExists(ctx context.Context) {
	err := w.redisClient.XGroupCreateMkStream(ctx, EMAIL_STREAM, EMAIL_GROUP, "$").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		slog.Error("Failed to create consumer group", "error", err)
		os.Exit(1) // Fatal equivalent
	}
}

func (w *EmailWorker) processNextBatch(ctx context.Context) {
	streams, err := w.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    EMAIL_GROUP,
		Consumer: EMAIL_CONSUMER,
		Streams:  []string{EMAIL_STREAM, ">"},
		Count:    1,
		Block:    0,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return
		}
		slog.Error("XReadGroup error", "error", err)
		return
	}

	for _, stream := range streams {
		w.processStream(ctx, stream)
	}
}

func (w *EmailWorker) processStream(ctx context.Context, stream redis.XStream) {
	for _, msg := range stream.Messages {
		if err := w.Handler(ctx, msg); err != nil {
			slog.Error("Handler failed for message", "message_id", msg.ID, "error", err)
			continue
		}

		if err := w.redisClient.XAck(ctx, EMAIL_STREAM, EMAIL_GROUP, msg.ID).Err(); err != nil {
			slog.Error("Failed to ACK message", "message_id", msg.ID, "error", err)
		}
	}
}
