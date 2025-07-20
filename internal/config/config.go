package config

import (
	"flag"
	"strings"

	"github.com/caarlos0/env"
)

// holds the configuration for theserver
type ServerConfig struct {
	Host          string `env:"ADDRESS"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	FPath         string `env:"FILE_STORAGE_PATH"`
	Restore       bool   `env:"RESTORE"`
	LogLevel      string `env:"LOG_LEVEL" envDefault:"debug"` // log level for the server
}

// holds the configuration for the agent
type AgentConfig struct {
	Host           string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"debug"` // log level for the agent
}

// GetServerConfig parses env variables, command line flags and returns the server configuration.
func GetServerConfig() (ServerConfig, error) {
	cfg := ServerConfig{}

	// set cli flags
	flag.StringVar(&cfg.Host, "a", "localhost:8080", "address of HTTP server")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "store interval in seconds")
	flag.StringVar(&cfg.FPath, "f", "./metrics_storage.json", "file path for storing metrics")
	flag.BoolVar(&cfg.Restore, "r", false, "restore metrics from file")

	// parse flags
	flag.Parse()

	// override with env variables if exists
	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	cfg.Host = strings.TrimPrefix(cfg.Host, "http://")

	return cfg, nil
}

// GetAgentConfig parses env variables,command line flags and returns the agent configuration.
func GetAgentConfig() (AgentConfig, error) {
	cfg := AgentConfig{}

	flag.StringVar(&cfg.Host, "a", ":8080", "address and port of the server")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "reporting interval in seconds")
	flag.IntVar(&cfg.PollInterval, "p", 2, "polling interval in seconds")
	flag.Parse()

	err := env.Parse(&cfg) // env variables can override the flags
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
