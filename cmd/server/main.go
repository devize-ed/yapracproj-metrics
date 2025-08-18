package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/handler"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/devize-ed/yapracproj-metrics.git/internal/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// get the server configuration from environment variables or command line flags
	cfg, err := config.GetServerConfig()
	if err != nil {
		return fmt.Errorf("failed to get server config: %w", err)
	}
	// initialize the logger with the specified log level
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.SafeSync()

	// Initialize the repository based on the configuration
	repository := handler.NewRepository(context.Background(), cfg.Repository)
	if err := repository.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}
	defer func() {
		if err := repository.Close(); err != nil {
			logger.Log.Errorf("failed to close repository: %w", err)
		}
	}()

	// create a context that listens for OS signals to shut down the server
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// create a new HTTP server with the configuration and handler
	h := handler.NewHandler(repository, cfg.Sign.Key)
	srv := server.NewServer(cfg, h)

	if err = srv.Serve(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
