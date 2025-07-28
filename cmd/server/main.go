package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/handler"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	st "github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	"github.com/devize-ed/yapracproj-metrics.git/internal/server"
)

func main() {
	if err := run(); err != nil {
		logger.Log.Fatal(err)
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
	defer logger.Log.Sync()

	logger.Log.Infof("Server config: interval=%d fpath=%s restore=%v host=%s",
		cfg.StoreInterval, cfg.FPath, cfg.Restore, cfg.Host)

	// create a new in-memory storage
	ms, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// initialize the database connection if a DSN is provided
	db, err := initDBConnection(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize db connection: %w", err)
	}
	defer db.Close()

	// create a context that listens for OS signals to shut down the server
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	// start the interval saver to periodically save metrics to the file
	ms.IntervalSaver(ctx, cfg.StoreInterval, cfg.FPath)

	// create a new HTTP server with the configuration and handler
	h := handler.NewHandler(ms, db)
	srv := server.NewServer(cfg, ms, h)

	if err = srv.Serve(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func initStorage(cfg config.ServerConfig) (*st.MemStorage, error) {
	ms := st.NewMemStorage(cfg.StoreInterval, cfg.FPath)

	if cfg.Restore {
		if err := ms.Load(cfg.FPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("load metrics: %w", err)
		}
	}
	return ms, nil
}

func initDBConnection(cfg config.ServerConfig) (*sql.DB, error) {
	logger.Log.Debugf("Connecting to database with DSN: %s", cfg.DatabaseDSN)
	// Open a database connection using loaded configuration
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open sql driver: %w", err)
	}

	// Ping the database to ensure the connection is established
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	logger.Log.Debug("Database connection established successfully")
	return db, nil
}
