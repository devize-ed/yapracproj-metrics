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
	cfg, err := config.GetAgentConfig() // Get agent configuration.
	if err != nil {
		return fmt.Errorf("failed to get agent config: %w", err)
	}
	// Initialize the logger with the configuration.
	logger, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Fatalf("failed to sync logger: %v", err)
		}
	}()

	// Log the agent start information.
	logger.Infof("Agent config: %+v", cfg)

	client := resty.New() // Initialize HTTP client.

	a := agent.NewAgent(client, cfg, logger) // Create a new agent instance.

	// Create a context that listens for OS signals to shut down the agent.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// Start the agent to collect and report metrics.
	if err := a.Run(ctx); err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	return nil
}
