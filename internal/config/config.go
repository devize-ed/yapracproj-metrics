package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

// holds the configuration for theserver
type ServerConfig struct {
	Host string `env:"ADDRESS"`
}

// holds the configuration for the agent
type AgentConfig struct {
	Host           string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
}

// GetServerConfig parses env variables, command line flags and returns the server configuration.
func GetServerConfig() (ServerConfig, error) {
	cfg := ServerConfig{}
	err := env.Parse(&cfg)
	if err != nil {
		return cfg, err
	}
	if cfg.Host == "" {
		log.Println("Using flag:", cfg.Host)
		flag.StringVar(&cfg.Host, "a", "localhost:8080", "address of HTTP server")
	}
	flag.Parse()
	return cfg, nil
}

// GetAgentConfig parses env variables,command line flags and returns the agent configuration.
func GetAgentConfig() (AgentConfig, error) {
	cfg := AgentConfig{}
	err := env.Parse(&cfg)
	if err != nil {
		return cfg, err
	}
	if cfg.Host == "" {
		log.Println("Using flag:", cfg.Host)
		flag.StringVar(&cfg.Host, "a", ":8080", "address and port of the server")
	}
	if cfg.PollInterval == 0 {
		log.Println("Using flag:", cfg.ReportInterval)
		flag.IntVar(&cfg.ReportInterval, "r", 10, "reporting interval in seconds")
	}
	if cfg.PollInterval == 0 {
		log.Println("Using flag:", cfg.PollInterval)
		flag.IntVar(&cfg.PollInterval, "p", 2, "polling interval in seconds")
	}
	flag.Parse()

	return cfg, nil
}
