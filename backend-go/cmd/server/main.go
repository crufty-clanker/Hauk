// Hauk Go Backend - HTTP server for location sharing.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hauk/hauk-go/internal/config"
	handlers "github.com/hauk/hauk-go/internal/handler"
	"github.com/hauk/hauk-go/internal/store/redis"
)

func main() {
	// Load configuration.
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Set up structured logging.
	logger := setupLogger()

	logger.Info("starting hauk-go backend",
		"version", config.BackendVersion,
		"port", cfg.ServerPort,
		"storage", cfg.StorageBackend,
	)

	// Connect to Redis store.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	store, err := redis.NewStore(ctx, cfg)
	cancel()
	if err != nil {
		logger.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	// Create HTTP handler.
	handler := handlers.New(store, cfg, logger)

	// Create HTTP server.
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		logger.Info("shutting down server...")
		ctx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("server shutdown error", "error", err)
		}
	}()

	// Start server.
	logger.Info("listening on", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func setupLogger() *slog.Logger {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}
