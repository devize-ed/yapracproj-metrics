// Package config provides configuration management for both server and agent components.
// It handles parsing of environment variables, command-line flags, and validation.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env"
	agent "github.com/devize-ed/yapracproj-metrics.git/internal/agent/config"
	audit "github.com/devize-ed/yapracproj-metrics.git/internal/audit/config"
	encryption "github.com/devize-ed/yapracproj-metrics.git/internal/encryption/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	db "github.com/devize-ed/yapracproj-metrics.git/internal/repository/db/config"
	fs "github.com/devize-ed/yapracproj-metrics.git/internal/repository/fstorage/config"
	sign "github.com/devize-ed/yapracproj-metrics.git/internal/sign/config"
)

type Config interface {
	ServerConfig | AgentConfig
}

// ServerConfig holds the configuration for the server.
type ServerConfig struct {
	Connection ServerConn                  `json:"connection"`
	Repository repository.RepositoryConfig `json:"repository"`
	Sign       sign.SignConfig             `json:"sign"`
	Audit      audit.AuditConfig           `json:"audit"`
	Encryption encryption.EncryptionConfig `json:"encryption"`
	LogLevel   string                      `env:"LOG_LEVEL" json:"log_level"` // Log level for the server.
}

// ServerConn holds server address configuration.
type ServerConn struct {
	Host string `env:"ADDRESS" json:"host"` // Address of the HTTP server.
}

// AgentConfig holds the configuration for the agent.
type AgentConfig struct {
	Connection AgentConn                   `json:"connection"`
	Agent      agent.AgentConfig           `json:"agent"`
	Sign       sign.SignConfig             `json:"sign"`
	Encryption encryption.EncryptionConfig `json:"encryption"`
	LogLevel   string                      `env:"LOG_LEVEL" json:"log_level"` // Log level for the agent.
}

// AgentConn holds agent connection configuration.
type AgentConn struct {
	Host string `env:"ADDRESS" json:"host"` // Address of the HTTP server.
}

// defaultServerConfig returns the baseline defaults used when neither file, flags nor env provide values.
func defaultServerConfig() ServerConfig {
	return ServerConfig{
		Connection: ServerConn{
			Host: "localhost:8080",
		},
		Repository: repository.RepositoryConfig{
			FSConfig: fs.FStorageConfig{
				StoreInterval: 0,
				FPath:         "",
				Restore:       false,
			},
			DBConfig: db.DBConfig{
				DatabaseDSN: "",
			},
		},
		Sign:       sign.SignConfig{},
		Audit:      audit.AuditConfig{},
		Encryption: encryption.EncryptionConfig{},
		LogLevel:   "",
	}
}

// defaultAgentConfig returns the baseline defaults used when neither file, flags nor env provide values.
func defaultAgentConfig() AgentConfig {
	return AgentConfig{
		Connection: AgentConn{
			Host: "localhost:8080",
		},
		Agent: agent.AgentConfig{
			PollInterval:   2,
			ReportInterval: 10,
			EnableGzip:     true,
			EnableTestGet:  false,
			RateLimit:      10,
		},
		Sign:       sign.SignConfig{},
		Encryption: encryption.EncryptionConfig{},
		LogLevel:   "",
	}
}

// scanConfigPath scans the command line arguments for the config file path.
func scanConfigPath(args []string) string {
	// scan the command line arguments
	for i := 0; i < len(args); i++ {
		a := args[i]
		// check if the argument has the config file path
		if strings.HasPrefix(a, "-c=") || strings.HasPrefix(a, "-config=") || strings.HasPrefix(a, "--config=") {
			// get the config file path
			if k := strings.IndexByte(a, '='); k != -1 {
				return a[k+1:]
			}
		}
		// check if the argument is the config file path
		if a == "-c" || a == "-config" || a == "--config" {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
	}
	return ""
}

// getConfigPath gets the config file path from the command line arguments or environment variables.
func getConfigPath() string {
	// scan the command line arguments for the config file path
	if p := scanConfigPath(os.Args[1:]); p != "" {
		return p
	}
	// check the path from the environment variable
	if p := os.Getenv("CONFIG"); p != "" {
		return p
	}
	return ""
}

// overrideStringFromEnv overrides the string from the environment variable.
func overrideStringFromEnv(dst *string, envName string) {
	if v, ok := os.LookupEnv(envName); ok {
		*dst = v
	}
}

// GetServerConfig gets the server configuration from the command line arguments or environment variables.
func GetServerConfig() (ServerConfig, error) {
	// start from defaults
	cfg := defaultServerConfig()

	// declare flags for the config file path to be in the help
	var configPathFlag string
	flag.StringVar(&configPathFlag, "c", "", "path to config file")
	flag.StringVar(&configPathFlag, "config", "", "path to config file")

	// get the config file path and load the configuration from the file if path is provided
	var p string
	loadedFromEnv := false
	if pa := scanConfigPath(os.Args[1:]); pa != "" {
		p = pa
	} else if pe := os.Getenv("CONFIG"); pe != "" {
		p = pe
		loadedFromEnv = true
	}
	if p != "" {
		var err error
		cfg, err = loadConfigFromFile(cfg, p)
		if err != nil {
			return cfg, fmt.Errorf("load config file %q: %w", p, err)
		}
	}

	// bind the server flags to the configuration
	bindServerFlags(&cfg) // defaults are from the cfg, collected from the file
	flag.Parse()

	// parse the server environment variables
	if err := parseServerEnvs(&cfg); err != nil {
		return cfg, fmt.Errorf("parse server envs: %w", err)
	}
	// Allow empty env override specifically when config path comes from env
	if loadedFromEnv {
		if v, ok := os.LookupEnv("DATABASE_DSN"); ok {
			cfg.Repository.DBConfig.DatabaseDSN = v
		}
	}

	// validate the server configuration
	if err := validateServerConfig(&cfg); err != nil {
		return cfg, fmt.Errorf("validate server config: %w", err)
	}
	return cfg, nil
}

// bindServerFlags binds the server flags to the configuration.
func bindServerFlags(cfg *ServerConfig) {
	flag.StringVar(&cfg.Connection.Host, "a", cfg.Connection.Host, "address of HTTP server")
	flag.IntVar(&cfg.Repository.FSConfig.StoreInterval, "i", cfg.Repository.FSConfig.StoreInterval, "store interval, s")
	flag.StringVar(&cfg.Repository.FSConfig.FPath, "f", cfg.Repository.FSConfig.FPath, "file storage path")
	flag.StringVar(&cfg.Repository.DBConfig.DatabaseDSN, "d", cfg.Repository.DBConfig.DatabaseDSN, "database DSN")
	flag.BoolVar(&cfg.Repository.FSConfig.Restore, "r", cfg.Repository.FSConfig.Restore, "restore on start")
	flag.StringVar(&cfg.Sign.Key, "k", cfg.Sign.Key, "sign key")
	flag.StringVar(&cfg.Encryption.CryptoKey, "crypto-key", cfg.Encryption.CryptoKey, "path to crypto key")
	flag.StringVar(&cfg.Audit.AuditFile, "audit-file", cfg.Audit.AuditFile, "audit file path")
	flag.StringVar(&cfg.Audit.AuditURL, "audit-url", cfg.Audit.AuditURL, "audit URL")
}

// parseServerEnvs parses the server environment variables.
func parseServerEnvs(cfg *ServerConfig) error {
	if err := env.Parse(&cfg.Connection); err != nil {
		return fmt.Errorf("parse server connection: %w", err)
	}
	if err := env.Parse(&cfg.Repository.FSConfig); err != nil {
		return fmt.Errorf("parse server repository file storage config: %w", err)
	}
	if err := env.Parse(&cfg.Repository.DBConfig); err != nil {
		return fmt.Errorf("parse server repository database config: %w", err)
	}
	if err := env.Parse(&cfg.Sign); err != nil {
		return fmt.Errorf("parse server sign config: %w", err)
	}
	if err := env.Parse(&cfg.Encryption); err != nil {
		return fmt.Errorf("parse server encryption config: %w", err)
	}
	if err := env.Parse(&cfg.Audit); err != nil {
		return fmt.Errorf("parse server audit config: %w", err)
	}
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("parse server config: %w", err)
	}

	return nil
}

// validateServerConfig validates the server configuration.
func validateServerConfig(cfg *ServerConfig) error {
	if cfg.Repository.FSConfig.StoreInterval < 0 {
		return fmt.Errorf("STORE_INTERVAL must be non-negative (got %d)", cfg.Repository.FSConfig.StoreInterval)
	}
	return nil
}

// GetAgentConfig gets the agent configuration from the command line arguments or environment variables.
func GetAgentConfig() (AgentConfig, error) {
	// start from defaults
	cfg := defaultAgentConfig()
	// declare flags for the config file path to be in the help
	var configPathFlag string
	flag.StringVar(&configPathFlag, "c", "", "path to config file")
	flag.StringVar(&configPathFlag, "config", "", "path to config file")

	// get the config file path and load the configuration from the file if path is provided
	if p := getConfigPath(); p != "" {
		var err error
		cfg, err = loadConfigFromFile(cfg, p)
		if err != nil {
			return cfg, fmt.Errorf("load config file %q: %w", p, err)
		}
	}

	// bind the agent flags to the configuration
	bindAgentFlags(&cfg)
	flag.Parse()

	// parse the agent environment variables
	if err := parseAgentEnvs(&cfg); err != nil {
		return cfg, fmt.Errorf("parse agent envs: %w", err)
	}

	// validate the agent configuration
	if err := validateAgentConfig(&cfg); err != nil {
		return cfg, fmt.Errorf("validate agent config: %w", err)
	}
	return cfg, nil
}

// bindAgentFlags binds the agent flags to the configuration.
func bindAgentFlags(cfg *AgentConfig) {
	flag.StringVar(&cfg.Connection.Host, "a", cfg.Connection.Host, "address of HTTP server")
	flag.IntVar(&cfg.Agent.ReportInterval, "r", cfg.Agent.ReportInterval, "reporting interval, s")
	flag.IntVar(&cfg.Agent.PollInterval, "p", cfg.Agent.PollInterval, "polling interval, s")
	flag.BoolVar(&cfg.Agent.EnableGzip, "gzip", cfg.Agent.EnableGzip, "enable gzip") // <— раньше было -c
	flag.BoolVar(&cfg.Agent.EnableTestGet, "g", cfg.Agent.EnableTestGet, "enable GET /metrics")
	flag.IntVar(&cfg.Agent.RateLimit, "l", cfg.Agent.RateLimit, "rate limit")
	flag.StringVar(&cfg.Sign.Key, "k", cfg.Sign.Key, "sign key")
	flag.StringVar(&cfg.Encryption.CryptoKey, "crypto-key", cfg.Encryption.CryptoKey, "path to crypto key")
}

// parseAgentEnvs parses the agent environment variables.
func parseAgentEnvs(cfg *AgentConfig) error {
	if err := env.Parse(&cfg.Connection); err != nil {
		return fmt.Errorf("parse agent connection: %w", err)
	}
	if err := env.Parse(&cfg.Agent); err != nil {
		return fmt.Errorf("parse agent config: %w", err)
	}
	if err := env.Parse(&cfg.Sign); err != nil {
		return fmt.Errorf("parse agent sign config: %w", err)
	}
	if err := env.Parse(&cfg.Encryption); err != nil {
		return fmt.Errorf("parse agent encryption config: %w", err)
	}
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("parse agent config: %w", err)
	}

	return nil
}

// validateAgentConfig validates the agent configuration.
func validateAgentConfig(cfg *AgentConfig) error {
	if cfg.Agent.PollInterval < 0 {
		return fmt.Errorf("POLL_INTERVAL must be non-negative (got %d)", cfg.Agent.PollInterval)
	}
	if cfg.Agent.ReportInterval < 0 {
		return fmt.Errorf("REPORT_INTERVAL must be non-negative (got %d)", cfg.Agent.ReportInterval)
	}
	if cfg.Agent.RateLimit < 1 {
		return fmt.Errorf("RATE_LIMIT must be greater than 0 (got %d)", cfg.Agent.RateLimit)
	}
	return nil
}

// loadConfigFromFile loads the configuration from a file.
func loadConfigFromFile[T Config](cfg T, path string) (T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config file: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal config: %w", err)
	}
	return cfg, nil
}

// GetBuildTag gets the build tag from the value.
func GetBuildTag(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}
