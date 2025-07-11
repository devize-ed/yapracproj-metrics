package config

import (
	"flag"
)

// holds the configuration for theserver
type ServerConfig struct {
	Host string
}

// holds the configuration for the agent
type AgentConfig struct {
	Host           string
	PollInterval   int
	ReportInterval int
}

// GetServerConfig parses command line flags and returns the server configuration.
func GetServerConfig() ServerConfig {
	cfg := ServerConfig{}

	flag.StringVar(&cfg.Host, "a", "localhost:8080", "address of HTTP server")
	flag.Parse()

	return cfg
}

// GetAgentConfig parses command line flags and returns the agent configuration.
func GetAgentConfig() AgentConfig {
	cfg := AgentConfig{}

	flag.StringVar(&cfg.Host, "a", ":8080", "address and port of the server")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "reporting interval in seconds")
	flag.IntVar(&cfg.PollInterval, "p", 2, "polling interval in seconds")
	flag.Parse()

	return cfg
}
