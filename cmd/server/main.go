package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	logger.Log.Infof("Server config: interval=%d fpath=%s restore=%v host=%s",
		cfg.StoreInterval, cfg.FPath, cfg.Restore, cfg.Host)

	// initialize the logger with the specified log level
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Log.Sync()

	// create a new in-memory storage
	ms := st.NewMemStorage(cfg.StoreInterval, cfg.FPath)
	if cfg.Restore {
		if err := ms.Load(cfg.FPath); err != nil {
			return fmt.Errorf("failed to load metrics: %w", err)
		}
	}

	// create a context that listens for OS signals to shut down the server
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	// start the interval saver to periodically save metrics to the file
	ms.IntervalSaver(ctx, cfg.StoreInterval, cfg.FPath)

	// create a new HTTP server with the configuration and handler
	h := handler.NewHandler(ms)
	srv := server.NewServer(cfg, h)

	// loging the address and starting the server
	go func() {
		logger.Log.Infof("HTTP server listening on %s", cfg.Host)
		if err := srv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Log.Errorf("listen error: %v", err)
		} else {
			logger.Log.Debug("HTTP server closed")
		}
	}()

	// wait for the signal to stop the server
	<-ctx.Done()
	logger.Log.Info("Stop signal received, shutting down the server...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Errorf("error shutting down the server: %v", err)
	}

	// save the metrics before exiting
	logger.Log.Debugf("Saving before exit ...")
	err = ms.Save(cfg.FPath)
	if err != nil {
		logger.Log.Errorf("failed to save on exit: %v", err)
	}

	return nil
}
