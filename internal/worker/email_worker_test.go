package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestEmailWorker_ProcessMessage(t *testing.T) {
	// Setup miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	// Create redis client pointing to miniredis
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create a channel to signal when a message is processed
	processed := make(chan string, 1)

	// Create worker with custom handler
	worker := NewEmailWorker(rdb, &MockMailService{})
	worker.Handler = func(ctx context.Context, msg redis.XMessage) error {
		processed <- msg.ID
		return nil
	}

	// Start worker in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.Start(ctx)

	// Give the worker a moment to start and create the group
	time.Sleep(100 * time.Millisecond)

	// Push a message to the stream
	id, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: EMAIL_STREAM,
		Values: map[string]interface{}{
			"to":      "test@example.com",
			"subject": "test subject",
			"body":    "hello world",
		},
	}).Result()
	assert.NoError(t, err)

	// Wait for processing
	select {
	case processedID := <-processed:
		assert.Equal(t, id, processedID)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for message processing")
	}

	// Verify ACK
	pending, err := rdb.XPending(ctx, EMAIL_STREAM, EMAIL_GROUP).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), pending.Count, "Message should be ACKed and removed from PEL")
}

func TestEmailWorker_ProcessMessage_Integration_Local(t *testing.T) {
	// Connection string for local Redis
	redisAddr := "localhost:6379"

	// Try to connect
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Skipping integration test: likely no local redis running: %v", err)
	}
	defer rdb.Close()

	// Use a unique stream name to avoid conflicts
	streamName := fmt.Sprintf("email-test-stream-%d", time.Now().UnixNano())
	groupName := fmt.Sprintf("email-test-group-%d", time.Now().UnixNano())

	// Create a channel to signal when a message is processed
	processed := make(chan string, 1)

	// Create worker with custom handler
	worker := NewEmailWorker(rdb, &MockMailService{})
	worker.Stream = streamName
	worker.Group = groupName
	worker.Consumer = "test-consumer"

	worker.Handler = func(ctx context.Context, msg redis.XMessage) error {
		processed <- msg.ID
		return nil
	}

	// Start worker in a goroutine
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	go worker.Start(workerCtx)

	// Give the worker a moment to start and create the group
	time.Sleep(100 * time.Millisecond)

	// Push a message to the stream
	id, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: map[string]interface{}{
			"to":      "integration@example.com",
			"subject": "integration subject",
			"body":    "hello integration",
		},
	}).Result()
	assert.NoError(t, err)

	// Wait for processing
	select {
	case processedID := <-processed:
		assert.Equal(t, id, processedID)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message processing")
	}

	// Verify ACK
	pending, err := rdb.XPending(ctx, streamName, groupName).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), pending.Count, "Message should be ACKed")

	// Cleanup
	rdb.Del(ctx, streamName)
	rdb.XGroupDestroy(ctx, streamName, groupName)
}

type MockMailService struct{}

func (m *MockMailService) SendEmail(to string, subject string, body string, messageId string) error {
	return nil
}
