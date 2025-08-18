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

// Server wraps an *http.Server and adds storage and configuration.
type Server struct {
	*http.Server
	cfg config.ServerConfig
}

// NewServer constructs the HTTP server using config and storage.
func NewServer(cfg config.ServerConfig, h *handler.Handler) *Server {
	s := &Server{
		cfg: cfg,
	}
	s.Server = &http.Server{
		Addr:    cfg.Connection.Host,
		Handler: h.NewRouter(),
	}
	return s
}

// Shutdown passes through to the embedded http.Server.
func (s *Server) shutdown(ctx context.Context) error {
	return s.Shutdown(ctx)
}

// Serve starts the HTTP server, blocks until ctx is cancelled, provide shutdown and save metrics.
func (s *Server) Serve(ctx context.Context) error {
	// Start the HTTP server in a goroutine.
	go func() {
		// Setart the server and listen for incoming requests.
		logger.Log.Infof("HTTP server listening on %s", s.cfg.Connection.Host)
		if err := s.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Log.Errorf("listen error: %w", err)
		} else {
			logger.Log.Debug("HTTP server closed")
		}
	}()

	// Wait for shutdown signal.
	<-ctx.Done()
	logger.Log.Info("Stop signal received, shutting down the server...")

	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.shutdown(shutCtx); err != nil {
		return fmt.Errorf("error shutting down the server: %w", err)
	}
	return nil
}
