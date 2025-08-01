package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/handler"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	st "github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	db "github.com/devize-ed/yapracproj-metrics.git/internal/repository/db"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository/fsaver"
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
	ms, err := initStorage(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	if closer, ok := ms.ExtStorage.(interface { Close () error }); ok {
		defer func () {
			if err := closer.Close(); err != nil {	
			}
		}()
	}

	// create a context that listens for OS signals to shut down the server
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// start the interval saver to periodically save metrics to the file
	ms.IntervalSaver(ctx, cfg.StoreInterval)

	// create a new HTTP server with the configuration and handler
	h := handler.NewHandler(ms)
	srv := server.NewServer(cfg, ms, h)

	if err = srv.Serve(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// initDBConnection initializes the database connection based on the provided configuration.
func initStorage(ctx context.Context, cfg config.ServerConfig) (*st.MemStorage, error) {
	// Initialize the storage based on the configuration
	var (
		stogare st.ExtStorage
		err     error
	)

	if cfg.DatabaseDSN != "" {
		stogare, err = db.NewDB(ctx, cfg.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database connection: %w", err)
		}
	} else if cfg.FPath != "" {
		stogare = fsaver.NewFileSaver(cfg.FPath)
	} else {
		stogare = st.NewStubStorage()
	}

	// Create a new in-memory storage
	ms := st.NewMemStorage(cfg.StoreInterval, stogare)

	// If restore is enabled, load the metrics from the repository
	if cfg.Restore {
		if err := ms.LoadFromRepo(ctx); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("load metrics: %w", err)
		}
	}
	return ms, nil
}
