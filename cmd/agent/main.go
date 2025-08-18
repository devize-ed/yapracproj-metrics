package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/devize-ed/yapracproj-metrics.git/internal/agent"
	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/go-resty/resty/v2"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}

}

func run() error {
	cfg, err := config.GetAgentConfig() // Get agent configuration.
	if err != nil {
		return fmt.Errorf("failed to get agent config: %w", err)
	}
	// Initialize the logger with the configuration.
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.SafeSync()

	// Log the agent start information.
	logger.Log.Infof("Agent config: %+v", cfg)

	client := resty.New() // Initialize HTTP client.

	a := agent.NewAgent(client, cfg) // Create a new agent instance.

	// Create a context that listens for OS signals to shut down the agent.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// Start the agent to collect and report metrics.
	if err := a.Run(ctx); err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	return nil
}
