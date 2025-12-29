package worker

import (
	"context"
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
	worker := NewEmailWorker(rdb)
	worker.Handler = func(ctx context.Context, msg redis.XMessage) error {
		processed <- msg.ID
		return nil
	}

	// Create stream and group (worker does this, but we can pre-populate if needed,
	// however worker.Start creates the group. We just need to ensure stream exists or is created by XAdd)

	// In the worker.Start, it creates the group with MkStream, so we don't need to pre-create.

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
			"email": "test@example.com",
			"body":  "hello world",
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
	// miniredis might not support XPending fully or we can check PEL?
	// Actually, let's just trust XAck was called if we want, or we can check pending messages count
	// If XAck is effective, the message should not be in the pending entry list (PEL) ideally,
	// or it should be removed from PEL (wait, XACK removes from PEL).
	// Let's check XInfoGroups or XPending
	// miniredis support for streams is good but let's verify.

	// XPending to check if there are any pending messages. Should be 0 if ACKed.
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
	// streamName := "email-test-stream-" + time.Now().Format("20060102150405")
	// groupName := "email-test-group"

	// Create a channel to signal when a message is processed
	processed := make(chan string, 1)

	// Create worker with custom handler
	worker := NewEmailWorker(rdb)

	// We need to override the stream/group constants in the worker for this test,
	// but they are constants.
	// Ideally we should refactor EmailWorker to accept config or stream names.
	// For now, since they are constants in the package, we can't easily change them without refactoring more.
	// However, since we are in the same package `service`, we can't change constants.
	// Wait, they are defined as:
	// const (
	// 	EMAIL_STREAM   = "go-email:microservice"
	// 	EMAIL_GROUP    = "go-email:group"
	// 	EMAIL_CONSUMER = "go-email:consumer"
	// )

	// If I cannot change them, I must use them.
	// To perform a clean test, I should ensure the stream/group is clean or handle existing data.
	// I'll stick to the constants but acknowledge this limits parallel testing or requires cleanup.

	// Let's use the constants but try to cleanup before/after.
	rdb.Del(ctx, EMAIL_STREAM)
	rdb.XGroupDestroy(ctx, EMAIL_STREAM, EMAIL_GROUP)

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
		Stream: EMAIL_STREAM,
		Values: map[string]interface{}{
			"email": "integration@example.com",
			"body":  "hello integration",
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
	pending, err := rdb.XPending(ctx, EMAIL_STREAM, EMAIL_GROUP).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), pending.Count, "Message should be ACKed")

	// Cleanup
	rdb.Del(ctx, EMAIL_STREAM)
	rdb.XGroupDestroy(ctx, EMAIL_STREAM, EMAIL_GROUP)
}
