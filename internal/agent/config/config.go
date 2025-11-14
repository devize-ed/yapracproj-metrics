// Package agent provides configuration structures for the agent component.
package agent

type AgentConfig struct {
	PollInterval   int  `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval int  `env:"REPORT_INTERVAL" json:"report_interval"`
	EnableGzip     bool `env:"ENABLE_GZIP" json:"enable_gzip"`               // Enable gzip compression for requests.
	EnableTestGet  bool `env:"ENABLE_GET_METRICS" json:"enable_get_metrics"` // Enable test retrieval of metrics from the server.
	RateLimit      int  `env:"RATE_LIMIT" json:"rate_limit"`
}
