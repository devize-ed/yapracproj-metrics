package main

import (
	"fmt"
	"log"

	"github.com/devize-ed/yapracproj-metrics.git/internal/agent"
	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		logger.Log.Fatal(err)
	}

}

func run() error {
	cfg, err := config.GetAgentConfig() // call the function to get agent configuration
	if err != nil {
		log.Fatal("Failed to get agent config:", err)
	}
	// Initialize the logger with the configuration
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Log.Sync()
	client := resty.New() // init client

	a := agent.NewAgent(client, cfg) // create a new agent instance
	a.Run()                          // start the agent to collect and report metrics

	// log the agent start information
	logger.Log.Info("Agent started with config:",
		zap.String("address", cfg.Host),
		zap.Int("poll_interval", cfg.PollInterval),
		zap.Int("report_interval", cfg.ReportInterval),
	)
	return nil
}
