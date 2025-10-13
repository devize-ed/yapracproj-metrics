package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/devize-ed/yapracproj-metrics.git/internal/audit"
	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/handler"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository"
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
	logger, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Errorf("failed to sync logger: %v", err)
		}
	}()

	// Initialize the repository based on the configuration
	repository, err := repository.NewRepository(context.Background(), cfg.Repository, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}
	if err := repository.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}
	defer func() {
		if err := repository.Close(); err != nil {
			logger.Errorf("failed to close repository: %w", err)
		}
	}()

	// create a context that listens for OS signals to shut down the server
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// create a new auditor with the logger
	auditor := audit.NewAuditor(logger)
	// start the auditor
	go auditor.Run(ctx)
	// if audit file is set, start the file auditor
	if cfg.Audit.AuditFile != "" {
		ch := auditor.Register()
		go audit.RunFileAudit(ctx, ch, cfg.Audit.AuditFile, logger)
	}
	if cfg.Audit.AuditURL != "" {
		// if audit URL is set, start the URL auditor
		ch := auditor.Register()
		go audit.RunURLAudit(ctx, ch, cfg.Audit.AuditURL, logger)
	}

	// create a new HTTP server with the configuration and handler
	h := handler.NewHandler(repository, cfg.Sign.Key, auditor, logger)
	srv := server.NewServer(cfg, h, logger)

	if err = srv.Serve(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
