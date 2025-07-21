package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/handler"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
)

// Repository matches the storage contract used by handler and mem‑storage.
type Repository interface {
	Save(path string) error
}

// Server wraps an *http.Server and adds storage and configuration.
type Server struct {
	*http.Server
	storage Repository
	cfg     config.ServerConfig
}

// NewServer constructs the HTTP server using config and storage.
func NewServer(cfg config.ServerConfig, storage Repository, h *handler.Handler) *Server {
	s := &Server{
		storage: storage,
		cfg:     cfg,
	}
	s.Server = &http.Server{
		Addr:    cfg.Host,
		Handler: h.NewRouter(),
	}
	return s
}

// Shutdown passes through to the embedded http.Server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}

// Serve starts the HTTP server, blocks until ctx is cancelled, provide shutdown and save metrics.
func (s *Server) Serve(ctx context.Context) error {
	// Start the HTTP server in a goroutine.
	go func() {
		logger.Log.Infof("HTTP server listening on %s", s.cfg.Host)
		if err := s.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Log.Errorf("listen error: %v", err)
		} else {
			logger.Log.Debug("HTTP server closed")
		}
	}()

	// Wait for shutdown signal.
	<-ctx.Done()
	logger.Log.Info("Stop signal received, shutting down the server...")

	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(shutCtx); err != nil {
		return fmt.Errorf("error shutting down the server: %w", err)
	}

	// Save metrics before the exit.
	logger.Log.Debug("Saving before exit ...")
	if err := s.storage.Save(s.cfg.FPath); err != nil {
		return fmt.Errorf("failed to save on exit: %w", err)
	}
	return nil
}
