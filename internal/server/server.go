// Package server provides HTTP server functionality.
// It handles server lifecycle, graceful shutdown, and request routing.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/handler"
	"go.uber.org/zap"
)

// Server wraps an *http.Server and adds storage and configuration.
type Server struct {
	*http.Server
	cfg    config.ServerConfig
	logger *zap.SugaredLogger
}

// NewServer constructs the HTTP server using config and storage.
func NewServer(cfg config.ServerConfig, h *handler.Handler, logger *zap.SugaredLogger) *Server {
	s := &Server{
		cfg:    cfg,
		logger: logger,
	}
	s.Server = &http.Server{
		Addr:    cfg.Connection.Host,
		Handler: h.NewRouter(),
	}
	return s
}

// shutdown gracefully shuts down the HTTP server.
func (s *Server) shutdown(ctx context.Context) error {
	return s.Shutdown(ctx)
}

// Serve starts the HTTP server, blocks until ctx is cancelled, provide shutdown and save metrics.
func (s *Server) Serve(ctx context.Context) error {
	// Start the HTTP server in a goroutine.
	go func() {
		// Setart the server and listen for incoming requests.
		s.logger.Infof("HTTP server listening on %s", s.cfg.Connection.Host)
		if err := s.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			s.logger.Errorf("listen error: %w", err)
		} else {
			s.logger.Debug("HTTP server closed")
		}
	}()

	// Wait for shutdown signal.
	<-ctx.Done()
	s.logger.Info("Stop signal received, shutting down the server...")

	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.shutdown(shutCtx); err != nil {
		return fmt.Errorf("error shutting down the server: %w", err)
	}
	return nil
}