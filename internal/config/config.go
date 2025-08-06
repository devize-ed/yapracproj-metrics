package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env"
	agent "github.com/devize-ed/yapracproj-metrics.git/internal/agent/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository"
)

// ServerConfig holds the configuration for the server.
type ServerConfig struct {
	Connection serverConn
	Repository repository.RepositoryConfig
	LogLevel   string `env:"LOG_LEVEL" envDefault:"debug"` // Log level for the server.
}

type serverConn struct {
	Host string `env:"ADDRESS"` // Address of the HTTP server.
}

// AgentConfig holds the configuration for the agent.
type AgentConfig struct {
	Connection agentConn
	Agent      agent.AgentConfig
	LogLevel   string `env:"LOG_LEVEL" envDefault:"debug"` // Log level for the agent.
}

type agentConn struct {
	Host string `env:"ADDRESS"` // Address of the HTTP server.
}

// GetServerConfig parses environment variables and command-line flags, then returns the server configuration.
func GetServerConfig() (ServerConfig, error) {
	cfg := ServerConfig{}

	// Set CLI flags.
	flag.StringVar(&cfg.Connection.Host, "a", "localhost:8080", "address of HTTP server")
	flag.IntVar(&cfg.Repository.FileConfig.StoreInterval, "i", 300, "store interval in seconds")
	flag.StringVar(&cfg.Repository.FileConfig.FPath, "f", "", "file path for storing metrics")
	flag.StringVar(&cfg.Repository.DBConfig.DatabaseDSN, "d", "", "string for the database connection")
	flag.BoolVar(&cfg.Repository.FileConfig.Restore, "r", false, "restore metrics from file")

	// Parse flags.
	flag.Parse()

	// Override with environment variables if they exist.
	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	// Validate the configuration.
	if cfg.Repository.FileConfig.StoreInterval < 0 {
		return cfg, fmt.Errorf("STORE_INTERVAL must be non-negative (got %d)", cfg.Repository.FileConfig.StoreInterval)
	}

	cfg.Connection.Host = strings.TrimPrefix(cfg.Connection.Host, "http://")

	return cfg, nil
}

// GetAgentConfig parses environment variables and command-line flags, then returns the agent configuration.
func GetAgentConfig() (AgentConfig, error) {
	cfg := AgentConfig{}

	flag.StringVar(&cfg.Connection.Host, "a", ":8080", "address and port of the server")
	flag.IntVar(&cfg.Agent.ReportInterval, "r", 10, "reporting interval in seconds")
	flag.IntVar(&cfg.Agent.PollInterval, "p", 2, "polling interval in seconds")
	flag.BoolVar(&cfg.Agent.EnableGzip, "c", true, "enable gzip compression for requests")
	flag.BoolVar(&cfg.Agent.EnableTestGet, "g", false, "enable test retrieval of metrics from the server")

	// Parse flags.
	flag.Parse()

	// Environment variables can override the flags.
	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	// Validate the configuration.
	if cfg.Agent.PollInterval < 0 {
		return cfg, fmt.Errorf("POLL_INTERVAL must be non-negative (got %d)", cfg.Agent.PollInterval)
	}
	if cfg.Agent.ReportInterval < 0 {
		return cfg, fmt.Errorf("REPORT_INTERVAL must be non-negative (got %d)", cfg.Agent.ReportInterval)
	}

	return cfg, nil
}
