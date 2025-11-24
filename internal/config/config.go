// Package config provides configuration management for both server and agent components.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	agent "github.com/devize-ed/yapracproj-metrics.git/internal/agent/config"
	audit "github.com/devize-ed/yapracproj-metrics.git/internal/audit/config"
	encryption "github.com/devize-ed/yapracproj-metrics.git/internal/encryption/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	db "github.com/devize-ed/yapracproj-metrics.git/internal/repository/db/config"
	fs "github.com/devize-ed/yapracproj-metrics.git/internal/repository/fstorage/config"
	sign "github.com/devize-ed/yapracproj-metrics.git/internal/sign/config"
)

// ServerConfig holds the configuration for the server.
type ServerConfig struct {
	Connection ServerConn                  `json:"connection"`
	Repository repository.RepositoryConfig `json:"repository"`
	Sign       sign.SignConfig             `json:"sign"`
	Audit      audit.AuditConfig           `json:"audit"`
	Encryption encryption.EncryptionConfig `json:"encryption"`
	LogLevel   string                      `json:"log_level"` // Log level for the server.
}

// ServerConn holds server address configuration.
type ServerConn struct {
	Host          string `env:"ADDRESS"`   // Address of the HTTP server.
	GRPCHost      string `env:"GRPC_HOST"` // Host of the gRPC server.
	TrustedSubnet string `env:"TRUSTED_SUBNET"`
}

// AgentConfig holds the configuration for the agent.
type AgentConfig struct {
	Connection      AgentConn                   `json:"connection"`
	Agent           agent.AgentConfig           `json:"agent"`
	Sign            sign.SignConfig             `json:"sign"`
	Encryption      encryption.EncryptionConfig `json:"encryption"`
	LogLevel        string                      `json:"log_level"`        // Log level for the agent.
	ShutdownTimeout int                         `json:"shutdown_timeout"` // Shutdown timeout for the agent.
}

// AgentConn holds agent connection configuration.
type AgentConn struct {
	Host     string `env:"ADDRESS"`   // Address of the HTTP server.
	GRPCHost string `env:"GRPC_HOST"` // Host of the gRPC server.
}

// defaultServerConfig returns the baseline defaults used when neither file, flags nor env provide values.
func defaultServerConfig() ServerConfig {
	return ServerConfig{
		Connection: ServerConn{
			Host:     "localhost:8080",
			GRPCHost: "localhost:3200",
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
			Host:     "localhost:8080",
			GRPCHost: "localhost:3200",
		},
		Agent: agent.AgentConfig{
			PollInterval:   2,
			ReportInterval: 10,
			EnableGzip:     true,
			EnableTestGet:  false,
			RateLimit:      10,
		},
		Sign:            sign.SignConfig{},
		Encryption:      encryption.EncryptionConfig{},
		ShutdownTimeout: 5,
		LogLevel:        "",
	}
}

// envBinding binds an environment variable to a viper keys
type envBinding struct {
	Key string // viper key
	Env string // env var
	Typ string
}

// bind environment variables to server config
var serverEnv = []envBinding{
	{"connection.host", "ADDRESS", "string"},
	{"connection.grpc_host", "GRPC_HOST", "string"},
	{"connection.trusted_subnet", "TRUSTED_SUBNET", "string"},
	{"repository.fs.store_interval", "STORE_INTERVAL", "int"},
	{"repository.fs.file_storage_path", "FILE_STORAGE_PATH", "string"},
	{"repository.fs.restore", "RESTORE", "bool"},
	{"repository.db.database_dsn", "DATABASE_DSN", "string"},
	{"sign.key", "KEY", "string"},
	{"encryption.crypto_key", "CRYPTO_KEY", "string"},
	{"audit.audit_file", "AUDIT_FILE", "string"},
	{"audit.audit_url", "AUDIT_URL", "string"},
	{"log_level", "LOG_LEVEL", "string"},
}

// bind environment variables to agent config
var agentEnv = []envBinding{
	{"connection.host", "ADDRESS", "string"},
	{"connection.grpc_host", "GRPC_HOST", "string"},
	{"agent.report_interval", "REPORT_INTERVAL", "int"},
	{"agent.poll_interval", "POLL_INTERVAL", "int"},
	{"agent.enable_gzip", "ENABLE_GZIP", "bool"},
	{"agent.enable_get_metrics", "ENABLE_TEST_GET", "bool"},
	{"agent.rate_limit", "RATE_LIMIT", "int"},
	{"sign.key", "KEY", "string"},
	{"encryption.crypto_key", "CRYPTO_KEY", "string"},
	{"log_level", "LOG_LEVEL", "string"},
	{"shutdown_timeout", "SHUTDOWN_TIMEOUT", "int"},
}

// bind environment variables to config
func bindEnvAndElevate(v *viper.Viper, bindings []envBinding) error {
	for _, b := range bindings {
		// if env is set and non-empty, set the value to the viper key
		if val, ok := os.LookupEnv(b.Env); ok && val != "" {
			switch b.Typ {
			case "string":
				v.Set(b.Key, val)
			case "int":
				n, err := strconv.Atoi(val)
				if err != nil {
					return fmt.Errorf("env %s: %w", b.Env, err)
				}
				v.Set(b.Key, n)
			case "bool":
				bv, err := strconv.ParseBool(val)
				if err != nil {
					return fmt.Errorf("env %s: %w", b.Env, err)
				}
				v.Set(b.Key, bv)
			}
		}
	}
	return nil
}

// mapServerFlagToKey maps server flag names to viper configuration keys.
func mapServerFlagToKey(flagName string) string {
	flagMap := map[string]string{
		"a":          "connection.host",
		"grpc-host":  "connection.grpc_host",
		"t":          "connection.trusted_subnet",
		"i":          "repository.fs.store_interval",
		"f":          "repository.fs.file_storage_path",
		"d":          "repository.db.database_dsn",
		"r":          "repository.fs.restore",
		"k":          "sign.key",
		"crypto-key": "encryption.crypto_key",
		"audit-file": "audit.audit_file",
		"audit-url":  "audit.audit_url",
	}
	if key, ok := flagMap[flagName]; ok {
		return key
	}
	return flagName
}

// mapAgentFlagToKey maps agent flag names to viper configuration keys.
func mapAgentFlagToKey(flagName string) string {
	flagMap := map[string]string{
		"a":          "connection.host",
		"grpc-host":  "connection.grpc_host",
		"r":          "agent.report_interval",
		"p":          "agent.poll_interval",
		"gzip":       "agent.enable_gzip",
		"g":          "agent.enable_get_metrics",
		"l":          "agent.rate_limit",
		"k":          "sign.key",
		"crypto-key": "encryption.crypto_key",
	}
	if key, ok := flagMap[flagName]; ok {
		return key
	}
	return flagName
}

// scanConfigPath scans the command line arguments for the config file path.
func scanConfigPath(args []string) string {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-c=") || strings.HasPrefix(a, "-config=") || strings.HasPrefix(a, "--config=") {
			if k := strings.IndexByte(a, '='); k != -1 {
				return a[k+1:]
			}
		}
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
	if p := scanConfigPath(os.Args[1:]); p != "" {
		return p
	}
	if p := os.Getenv("CONFIG"); p != "" {
		return p
	}
	return ""
}

// setServerDefaults sets the default values for the server config.
func setServerDefaults(v *viper.Viper) {
	// get the default values
	d := defaultServerConfig()
	// set the default values
	v.SetDefault("connection.host", d.Connection.Host)
	v.SetDefault("connection.grpc_host", d.Connection.GRPCHost)
	v.SetDefault("connection.trusted_subnet", d.Connection.TrustedSubnet)
	v.SetDefault("repository.fs.store_interval", d.Repository.FSConfig.StoreInterval)
	v.SetDefault("repository.fs.file_storage_path", d.Repository.FSConfig.FPath)
	v.SetDefault("repository.fs.restore", d.Repository.FSConfig.Restore)
	v.SetDefault("repository.db.database_dsn", d.Repository.DBConfig.DatabaseDSN)
	v.SetDefault("sign.key", d.Sign.Key)
	v.SetDefault("encryption.crypto_key", d.Encryption.CryptoKey)
	v.SetDefault("audit.audit_file", d.Audit.AuditFile)
	v.SetDefault("audit.audit_url", d.Audit.AuditURL)
	v.SetDefault("log_level", d.LogLevel)
}

// setAgentDefaults sets the default values for the agent config.
func setAgentDefaults(v *viper.Viper) {
	// get the default values
	d := defaultAgentConfig()
	// set the default values
	v.SetDefault("connection.host", d.Connection.Host)
	v.SetDefault("connection.grpc_host", d.Connection.GRPCHost)
	v.SetDefault("agent.report_interval", d.Agent.ReportInterval)
	v.SetDefault("agent.poll_interval", d.Agent.PollInterval)
	v.SetDefault("agent.enable_gzip", d.Agent.EnableGzip)
	v.SetDefault("agent.enable_get_metrics", d.Agent.EnableTestGet)
	v.SetDefault("agent.rate_limit", d.Agent.RateLimit)
	v.SetDefault("sign.key", d.Sign.Key)
	v.SetDefault("encryption.crypto_key", d.Encryption.CryptoKey)
	v.SetDefault("log_level", d.LogLevel)
	v.SetDefault("shutdown_timeout", d.ShutdownTimeout)
}

// GetServerConfig gets the server config from the command line arguments or environment variables.
func GetServerConfig() (ServerConfig, error) {
	var cfg ServerConfig

	// create a new viper instance
	v := viper.New()
	// set the default values
	setServerDefaults(v)

	// Track which keys were explicitly set
	explicitlySet := make(map[string]bool)

	// read the config file if provided
	configFileKeys := make(map[string]bool)
	if p := getConfigPath(); p != "" {
		// set the config file path
		v.SetConfigFile(p)
		// read the config file
		if err := v.ReadInConfig(); err != nil {
			return cfg, fmt.Errorf("load config file %q: %w", p, err)
		}
		// Track keys that were set in the config file
		for _, key := range v.AllKeys() {
			if v.IsSet(key) {
				configFileKeys[key] = true
			}
		}
	}

	// create a new flag set
	fs := pflag.NewFlagSet("server", pflag.ContinueOnError)
	// add a flag for the config file path
	fs.StringP("config", "c", "", "path to config file")

	fs.StringP("a", "a", v.GetString("connection.host"), "address of HTTP server")
	fs.String("grpc-host", v.GetString("connection.grpc_host"), "address of gRPC server")
	fs.StringP("t", "t", v.GetString("connection.trusted_subnet"), "trusted subnet for IP filtering")
	fs.IntP("i", "i", v.GetInt("repository.fs.store_interval"), "store interval, s")
	fs.StringP("f", "f", v.GetString("repository.fs.file_storage_path"), "file storage path")
	fs.StringP("d", "d", v.GetString("repository.db.database_dsn"), "database DSN")
	fs.BoolP("r", "r", v.GetBool("repository.fs.restore"), "restore on start")
	fs.StringP("k", "k", v.GetString("sign.key"), "sign key")
	fs.String("crypto-key", v.GetString("encryption.crypto_key"), "path to crypto key")
	fs.String("audit-file", v.GetString("audit.audit_file"), "audit file path")
	fs.String("audit-url", v.GetString("audit.audit_url"), "audit URL")

	// Parse flags
	if err := fs.Parse(os.Args[1:]); err != nil && err != pflag.ErrHelp {
		return cfg, fmt.Errorf("parse server flags: %w", err)
	}

	// check if the flag was explicitly set
	fs.Visit(func(f *pflag.Flag) {
		key := mapServerFlagToKey(f.Name)
		explicitlySet[key] = true
		switch f.Value.Type() {
		case "string":
			v.Set(key, f.Value.String())
		case "int":
			if val, err := fs.GetInt(f.Name); err == nil {
				v.Set(key, val)
			}
		case "bool":
			if val, err := fs.GetBool(f.Name); err == nil {
				v.Set(key, val)
			}
		}
	})

	// bind the environment variables to the viper instance
	loadedFromEnv := os.Getenv("CONFIG") != ""
	if err := bindEnvAndElevate(v, serverEnv); err != nil {
		return cfg, fmt.Errorf("parse server envs: %w", err)
	}
	// Track env vars that were set
	for _, b := range serverEnv {
		if val, ok := os.LookupEnv(b.Env); ok && val != "" {
			explicitlySet[b.Key] = true
		}
	}
	// Special case: allow empty DATABASE_DSN override when CONFIG comes from env
	if loadedFromEnv {
		if dsnVal, ok := os.LookupEnv("DATABASE_DSN"); ok {
			v.Set("repository.db.database_dsn", dsnVal)
			explicitlySet["repository.db.database_dsn"] = true
		}
	}

	// unmarshal the viper instance into the server config
	decoderOpt := viper.DecoderConfigOption(func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	})
	if err := v.Unmarshal(&cfg, decoderOpt); err != nil {
		return cfg, fmt.Errorf("unmarshal server config: %w", err)
	}

	// Apply defaults for fields that weren't set explicitly (not in file, flags, or env)
	if cfg.Connection.GRPCHost == "" && !explicitlySet["connection.grpc_host"] && !configFileKeys["connection.grpc_host"] {
		cfg.Connection.GRPCHost = v.GetString("connection.grpc_host")
	}
	if cfg.Connection.TrustedSubnet == "" && !explicitlySet["connection.trusted_subnet"] && !configFileKeys["connection.trusted_subnet"] {
		cfg.Connection.TrustedSubnet = v.GetString("connection.trusted_subnet")
	}

	// validate the server config
	if err := validateServerConfig(&cfg); err != nil {
		return cfg, fmt.Errorf("validate server config: %w", err)
	}
	return cfg, nil
}

// validateServerConfig validates the server config.
func validateServerConfig(cfg *ServerConfig) error {
	if cfg.Repository.FSConfig.StoreInterval < 0 {
		return fmt.Errorf("STORE_INTERVAL must be non-negative (got %d)", cfg.Repository.FSConfig.StoreInterval)
	}
	return nil
}

// GetAgentConfig gets the agent config from the command line arguments or environment variables.
func GetAgentConfig() (AgentConfig, error) {
	var cfg AgentConfig

	// create a new viper instance
	v := viper.New()
	// set the default values
	setAgentDefaults(v)

	// Track which keys were explicitly set
	explicitlySet := make(map[string]bool)

	// read the config file if provided
	configFileKeys := make(map[string]bool)
	if p := getConfigPath(); p != "" {
		// set the config file path
		v.SetConfigFile(p)
		// read the config file
		if err := v.ReadInConfig(); err != nil {
			return cfg, fmt.Errorf("load config file %q: %w", p, err)
		}
		// Track keys that were set in the config file
		for _, key := range v.AllKeys() {
			if v.IsSet(key) {
				configFileKeys[key] = true
			}
		}
	}

	// create a new flag set
	fs := pflag.NewFlagSet("agent", pflag.ContinueOnError)
	// add a flag for the config file path
	fs.StringP("config", "c", "", "path to config file")

	fs.StringP("a", "a", v.GetString("connection.host"), "address of HTTP server")
	fs.String("grpc-host", v.GetString("connection.grpc_host"), "address of gRPC server")
	fs.IntP("r", "r", v.GetInt("agent.report_interval"), "reporting interval, s")
	fs.IntP("p", "p", v.GetInt("agent.poll_interval"), "polling interval, s")
	fs.Bool("gzip", v.GetBool("agent.enable_gzip"), "enable gzip")
	fs.BoolP("g", "g", v.GetBool("agent.enable_get_metrics"), "enable GET /metrics")
	fs.IntP("l", "l", v.GetInt("agent.rate_limit"), "rate limit")
	fs.StringP("k", "k", v.GetString("sign.key"), "sign key")
	fs.String("crypto-key", v.GetString("encryption.crypto_key"), "path to crypto key")

	// Parse flags
	if err := fs.Parse(os.Args[1:]); err != nil && err != pflag.ErrHelp {
		return cfg, fmt.Errorf("parse agent flags: %w", err)
	}

	// check if the flag was explicitly set
	fs.Visit(func(f *pflag.Flag) {
		key := mapAgentFlagToKey(f.Name)
		explicitlySet[key] = true
		switch f.Value.Type() {
		case "string":
			v.Set(key, f.Value.String())
		case "int":
			if val, err := fs.GetInt(f.Name); err == nil {
				v.Set(key, val)
			}
		case "bool":
			if val, err := fs.GetBool(f.Name); err == nil {
				v.Set(key, val)
			}
		}
	})

	// bind the environment variables to the viper instance
	if err := bindEnvAndElevate(v, agentEnv); err != nil {
		return cfg, fmt.Errorf("parse agent envs: %w", err)
	}
	// Track env vars that were set
	for _, b := range agentEnv {
		if val, ok := os.LookupEnv(b.Env); ok && val != "" {
			explicitlySet[b.Key] = true
		}
	}

	// unmarshal the viper instance into the agent config
	decoderOpt := viper.DecoderConfigOption(func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	})
	if err := v.Unmarshal(&cfg, decoderOpt); err != nil {
		return cfg, fmt.Errorf("unmarshal agent: %w", err)
	}

	// Apply defaults for fields that weren't set explicitly (not in file, flags, or env)
	if cfg.Connection.GRPCHost == "" && !explicitlySet["connection.grpc_host"] && !configFileKeys["connection.grpc_host"] {
		cfg.Connection.GRPCHost = v.GetString("connection.grpc_host")
	}

	// validate the agent config
	if err := validateAgentConfig(&cfg); err != nil {
		return cfg, fmt.Errorf("validate agent config: %w", err)
	}
	return cfg, nil
}

// validateAgentConfig validates the agent config.
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
	if cfg.ShutdownTimeout < 0 {
		return fmt.Errorf("SHUTDOWN_TIMEOUT must be non-negative (got %d)", cfg.ShutdownTimeout)
	}
	return nil
}

func GetBuildTag(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}
