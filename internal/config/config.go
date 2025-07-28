package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env"
)

// ServerConfig holds the configuration for the server.
type ServerConfig struct {
	Host          string `env:"ADDRESS"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	FPath         string `env:"FILE_STORAGE_PATH"`
	Restore       bool   `env:"RESTORE"`
	DatabaseDSN   string `env:"DATABASE_DSN"`                 // string for the database connection.
	LogLevel      string `env:"LOG_LEVEL" envDefault:"debug"` // Log level for the server.
}

// AgentConfig holds the configuration for the agent.
type AgentConfig struct {
	Host           string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"debug"` // Log level for the agent.
	EnableGzip     bool   `env:"ENABLE_GZIP"`                  // Enable gzip compression for requests.
	EnableTestGet  bool   `env:"ENABLE_GET_METRICS"`           // Enable test retrieval of metrics from the server.
}

// GetServerConfig parses environment variables and command-line flags, then returns the server configuration.
func GetServerConfig() (ServerConfig, error) {
	cfg := ServerConfig{}

	// Set CLI flags.
	flag.StringVar(&cfg.Host, "a", "localhost:8080", "address of HTTP server")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "store interval in seconds")
	flag.StringVar(&cfg.FPath, "f", "./metrics_storage.json", "file path for storing metrics")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "string for the database connection")
	flag.BoolVar(&cfg.Restore, "r", false, "restore metrics from file")

	// Parse flags.
	flag.Parse()

	// Override with environment variables if they exist.
	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	// Validate the configuration.
	if cfg.StoreInterval < 0 {
		return cfg, fmt.Errorf("STORE_INTERVAL must be non-negative (got %d)", cfg.StoreInterval)
	}

	cfg.Host = strings.TrimPrefix(cfg.Host, "http://")

	return cfg, nil
}

// GetAgentConfig parses environment variables and command-line flags, then returns the agent configuration.
func GetAgentConfig() (AgentConfig, error) {
	cfg := AgentConfig{}

	flag.StringVar(&cfg.Host, "a", ":8080", "address and port of the server")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "reporting interval in seconds")
	flag.IntVar(&cfg.PollInterval, "p", 2, "polling interval in seconds")
	flag.BoolVar(&cfg.EnableGzip, "c", true, "enable gzip compression for requests")
	flag.BoolVar(&cfg.EnableTestGet, "g", false, "enable test retrieval of metrics from the server")

	// Parse flags.
	flag.Parse()

	// Environment variables can override the flags.
	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	// Validate the configuration.
	if cfg.PollInterval < 0 {
		return cfg, fmt.Errorf("POLL_INTERVAL must be non-negative (got %d)", cfg.PollInterval)
	}
	if cfg.ReportInterval < 0 {
		return cfg, fmt.Errorf("REPORT_INTERVAL must be non-negative (got %d)", cfg.ReportInterval)
	}

	return cfg, nil
}
