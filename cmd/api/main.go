package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kantha2004/go-mail-service/internal/config"
	"github.com/Kantha2004/go-mail-service/internal/logger"
	"github.com/Kantha2004/go-mail-service/internal/repository"
	"github.com/Kantha2004/go-mail-service/internal/service"
	"github.com/Kantha2004/go-mail-service/internal/worker"

	"github.com/lmittmann/tint"
)

func main() {

	// File handler (JSON)
	logFile, err := os.OpenFile("logs/server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	fileHandler := slog.NewJSONHandler(logFile, nil)

	// Console handler (Tint/Color)
	consoleHandler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
	})

	// Fanout handler
	logger := slog.New(logger.NewFanoutHandler(consoleHandler, fileHandler))
	slog.SetDefault(logger)

	config := config.LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	redisClient, err := repository.NewRedisClient(ctx, config.RedisAddr)

	if err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}

	mailService := service.NewMailTrapService(config.MailtrapAPIKey, config.MailtrapURL)
	emailWorker := worker.NewEmailWorker(redisClient, mailService)
	go emailWorker.Start(ctx)

	slog.Info("Starting Email Service", "port", config.ServerPort)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr: config.ServerPort,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")
	slog.Warn("Press Ctrl+C again to force quit")

	// Force shutdown on second signal
	go func() {
		<-quit
		slog.Info("Force quitting received...")
		os.Exit(1)
	}()

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTimeout()

	if err := server.Shutdown(ctxTimeout); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	// Cancel context to stop worker
	cancel()
	// Wait a bit for worker to finish if needed, or rely on worker.Stop() if it blocks
	// But here we just stop the worker.
	emailWorker.Stop()

	slog.Info("Server exiting")
}
