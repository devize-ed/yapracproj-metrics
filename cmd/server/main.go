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
	grpc "github.com/devize-ed/yapracproj-metrics.git/internal/grpc/server"
	"github.com/devize-ed/yapracproj-metrics.git/internal/handler"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	"github.com/devize-ed/yapracproj-metrics.git/internal/server"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	fmt.Printf("Version: %s, Build Date: %s, Build Commit: %s", config.GetBuildTag(buildVersion), config.GetBuildTag(buildDate), config.GetBuildTag(buildCommit))

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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	// create a new auditor with the logger
	auditor := audit.NewAuditor(logger, cfg.Audit.AuditFile, cfg.Audit.AuditURL)
	// start the auditor
	go auditor.Run(ctx)

	// create a new HTTP server with the configuration and handler
	h := handler.NewHandler(repository, cfg.Sign.Key, auditor, cfg.Connection.TrustedSubnet, logger)
	srv := server.NewServer(cfg, h, logger)

	// start the gRPC server
	grpcServer := grpc.NewServer(repository, logger)
	go grpcServer.Serve(ctx, cfg.Connection.GRPCHost, cfg.Connection.TrustedSubnet)

	if err = srv.Serve(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
