// Package config provides configuration management for both server and agent components.
// It handles parsing of environment variables, command-line flags, and validation.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env"
	agent "github.com/devize-ed/yapracproj-metrics.git/internal/agent/config"
	audit "github.com/devize-ed/yapracproj-metrics.git/internal/audit/config"
	encryption "github.com/devize-ed/yapracproj-metrics.git/internal/encryption/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	sign "github.com/devize-ed/yapracproj-metrics.git/internal/sign/config"
)

// ServerConfig holds the configuration for the server.
type ServerConfig struct {
	Connection ServerConn
	Repository repository.RepositoryConfig
	Sign       sign.SignConfig
	Audit      audit.AuditConfig
	Encryption encryption.EncryptionConfig
	LogLevel   string `env:"LOG_LEVEL" envDefault:"debug"` // Log level for the server.
	Config     string `env:"CONFIG"`                       // Path to the configuration file.
}

// ServerConn holds server address configuration.
type ServerConn struct {
	Host string `env:"ADDRESS"` // Address of the HTTP server.
}

// AgentConfig holds the configuration for the agent.
type AgentConfig struct {
	Connection AgentConn
	Agent      agent.AgentConfig
	Sign       sign.SignConfig
	Encryption encryption.EncryptionConfig
	LogLevel   string `env:"LOG_LEVEL" envDefault:"debug"` // Log level for the agent.
	Config     string `env:"CONFIG"`                       // Path to the configuration file.
}

// AgentConn holds agent connection configuration.
type AgentConn struct {
	Host string `env:"ADDRESS"` // Address of the HTTP server.
}

// GetServerConfig parses environment variables and command-line flags, then returns the server configuration.
func GetServerConfig() (ServerConfig, error) {
	cfg := ServerConfig{}

	// Check if flag or environment variable for config file is provided
	if cfg.Config != "" || flag.Lookup("c") != nil {
		// load config from file
		cfg, err := json.Unmarshal([]byte(cfg.Config), &cfg)
		if err != nil {
			return cfg, fmt.Errorf("failed to unmarshal config file: %w", err)
		}
		return cfg, nil
	}

	// Set CLI flags.
	flag.StringVar(&cfg.Connection.Host, "a", "localhost:8080", "address of HTTP server")
	flag.IntVar(&cfg.Repository.FSConfig.StoreInterval, "i", 300, "store interval in seconds")
	flag.StringVar(&cfg.Repository.FSConfig.FPath, "f", "", "file path for storing metrics")
	flag.StringVar(&cfg.Repository.DBConfig.DatabaseDSN, "d", "", "string for the database connection")
	flag.BoolVar(&cfg.Repository.FSConfig.Restore, "r", false, "restore metrics from file")
	flag.StringVar(&cfg.Sign.Key, "k", "", "secret key for the Hash")
	flag.StringVar(&cfg.Encryption.CryptoKey, "crypto-key", "", "path to the crypto key for the encryption")
	flag.StringVar(&cfg.Audit.AuditFile, "audit-file", "", "file path for storing audit")
	flag.StringVar(&cfg.Audit.AuditURL, "audit-url", "", "URL for storing audit")

	// Parse flags.
	flag.Parse()

	// Override with environment variables if they exist.
	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}
	if err := env.Parse(&cfg.Connection); err != nil {
		return cfg, err
	}
	if err := env.Parse(&cfg.Repository.FSConfig); err != nil {
		return cfg, err
	}
	if err := env.Parse(&cfg.Repository.DBConfig); err != nil {
		return cfg, err
	}
	if err := env.Parse(&cfg.Sign); err != nil {
		return cfg, err
	}
	if err := env.Parse(&cfg.Audit); err != nil {
		return cfg, err
	}
	if err := env.Parse(&cfg.Encryption); err != nil {
		return cfg, err
	}

	// Validate the configuration.
	if cfg.Repository.FSConfig.StoreInterval < 0 {
		return cfg, fmt.Errorf("STORE_INTERVAL must be non-negative (got %d)", cfg.Repository.FSConfig.StoreInterval)
	}

	cfg.Connection.Host = strings.TrimPrefix(cfg.Connection.Host, "http://")

	return cfg, nil
}

// GetAgentConfig parses environment variables and command-line flags, then returns the agent configuration.
func GetAgentConfig() (AgentConfig, error) {
	cfg := AgentConfig{}

	flag.StringVar(&cfg.Connection.Host, "a", "localhost:8080", "address and port of the server")
	flag.IntVar(&cfg.Agent.ReportInterval, "r", 10, "reporting interval in seconds")
	flag.IntVar(&cfg.Agent.PollInterval, "p", 2, "polling interval in seconds")
	flag.BoolVar(&cfg.Agent.EnableGzip, "c", true, "enable gzip compression for requests")
	flag.BoolVar(&cfg.Agent.EnableTestGet, "g", false, "enable test retrieval of metrics from the server")
	flag.IntVar(&cfg.Agent.RateLimit, "l", 1, "rate limit for the agent")
	flag.StringVar(&cfg.Sign.Key, "k", "", "secret key for the Hash")
	flag.StringVar(&cfg.Encryption.CryptoKey, "crypto-key", "", "path to the crypto key for the encryption")

	// Parse flags.
	flag.Parse()

	// Override with environment variables if they exist.
	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}
	if err := env.Parse(&cfg.Connection); err != nil {
		return cfg, err
	}
	if err := env.Parse(&cfg.Agent); err != nil {
		return cfg, err
	}
	if err := env.Parse(&cfg.Sign); err != nil {
		return cfg, err
	}
	if err := env.Parse(&cfg.Encryption); err != nil {
		return cfg, err
	}
	// Validate the configuration.
	if cfg.Agent.PollInterval < 0 {
		return cfg, fmt.Errorf("POLL_INTERVAL must be non-negative (got %d)", cfg.Agent.PollInterval)
	}
	if cfg.Agent.ReportInterval < 0 {
		return cfg, fmt.Errorf("REPORT_INTERVAL must be non-negative (got %d)", cfg.Agent.ReportInterval)
	}
	if cfg.Agent.RateLimit < 1 {
		return cfg, fmt.Errorf("RATE_LIMIT must be greater than 0 (got %d)", cfg.Agent.RateLimit)
	}
	return cfg, nil
}

func GetBuildTag(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}
