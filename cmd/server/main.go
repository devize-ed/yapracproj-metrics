package main

import (
	"fmt"

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
	ms := st.NewMemStorage() // init the memory storage for metrics
	h := handler.NewHandler(ms)
	cfg, err := config.GetServerConfig() // call the function to get server configuration
	if err != nil {
		return fmt.Errorf("failed to get server config: %w", err)
	}
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Log.Sync()
	srv := server.NewServer(cfg, h) // create a new server with the configuration and handler

	// loging the address and starting the server
	logger.Log.Info("Starting server on ", cfg.Host)
	err = srv.ListenAndServe()

	return fmt.Errorf("failed to start HTTP server: %w", err)
}
