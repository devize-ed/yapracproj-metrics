package main

import (
	"fmt"
	"log"

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
	defer logger.Log.Sync()

	// Log the agent start information.
	logger.Log.Infof("Agent config: poll_interval=%d report_interval=%d host=%s enable_gzip=%v",
		cfg.Agent.PollInterval, cfg.Agent.ReportInterval, cfg.Connection.Host, cfg.Agent.EnableGzip)

	client := resty.New() // Initialize HTTP client.

	a := agent.NewAgent(client, cfg) // Create a new agent instance.

	// Start the agent to collect and report metrics.
	if err := a.Run(); err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	return nil
}
