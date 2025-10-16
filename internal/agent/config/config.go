// Package agent provides configuration structures for the agent component.
package agent

type AgentConfig struct {
	PollInterval   int  `env:"POLL_INTERVAL"`
	ReportInterval int  `env:"REPORT_INTERVAL"`
	EnableGzip     bool `env:"ENABLE_GZIP"`        // Enable gzip compression for requests.
	EnableTestGet  bool `env:"ENABLE_GET_METRICS"` // Enable test retrieval of metrics from the server.
	RateLimit      int  `env:"RATE_LIMIT"`
}
