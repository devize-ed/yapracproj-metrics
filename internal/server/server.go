package server

import (
	"net/http"

	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/handler"
)

// create a new HTTP server with the given configuration and handler.
func NewServer(cfg config.ServerConfig, h *handler.Handler) *http.Server {
	return &http.Server{
		Addr:    cfg.Host,
		Handler: h.NewRouter(),
	}

}
