package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/go-resty/resty/v2"
)

// Agent holds the HTTP client, metric storage, and configuration.
type Agent struct {
	client  *resty.Client
	storage *AgentStorage
	config  config.AgentConfig
}

// NewAgent returns a new Agent that uses the given HTTP client and configuration.
func NewAgent(client *resty.Client, config config.AgentConfig) *Agent {
	return &Agent{
		client:  client,
		storage: NewAgentStorage(),
		config:  config,
	}
}

func (a *Agent) Run() error {
	// Convert interval values to time.Duration.
	timePollInterval := time.Duration(a.config.PollInterval) * time.Second
	timeReportInterval := time.Duration(a.config.ReportInterval) * time.Second

	// Set up the tickers for polling and reporting.
	pollTicker := time.NewTicker(timePollInterval)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(timeReportInterval)
	defer reportTicker.Stop()

	// Start the agent loop.
	for {
		select {
		case <-pollTicker.C: // Collect metrics at the polling interval.
			a.storage.CollectMetrics()
		case <-reportTicker.C: // Send metrics at the reporting interval.
			logger.Log.Debug("Reporting metrics...")

			// Check whether “test‑get” mode is enabled.
			if !a.config.EnableTestGet {
				// Iterate over the storage and send metrics to the server.
				for name, val := range a.storage.Counters {
					if err := SendMetric(a, name, val); err != nil {
						logger.Log.Error("error sending ", name, ": ", err)
					}
				}
				for name, val := range a.storage.Gauges {
					if err := SendMetric(a, name, val); err != nil {
						logger.Log.Error("error sending ", name, ": ", err)
					}
				}
			} else {
				// “Test‑get” mode: request metrics from the server.
				for name, val := range a.storage.Counters {
					if err := GetMetric(a, name, val); err != nil {
						logger.Log.Error("error getting ", name, ": ", err)
					}
				}
				for name, val := range a.storage.Gauges {
					if err := GetMetric(a, name, val); err != nil {
						logger.Log.Error("error getting ", name, ": ", err)
					}
				}
			}
		}
	}
}

// SendMetric sends a single metric to the server.
func SendMetric[T MetricValue](a *Agent, metric string, value T) error {
	endpoint := fmt.Sprintf("http://%s/value/", a.config.Host)
	body := models.Metrics{
		ID: metric,
	}

	// Set the value and metric type.
	switch v := any(value).(type) {
	case Gauge:
		body.MType = models.Gauge
		floatValue := float64(v)
		body.Value = &floatValue
	case Counter:
		body.MType = models.Counter
		intValue := int64(v)
		body.Delta = &intValue
	default:
		return fmt.Errorf("unsupported metric type %T", v)
	}

	err := a.Request(metric, endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to send: %v", err)
	}
	return nil
}

// GetMetric requests a metric from the server for testing purposes.
func GetMetric[T MetricValue](a *Agent, metric string, value T) error {
	endpoint := fmt.Sprintf("http://%s/value/", a.config.Host)

	body := models.Metrics{
		ID: metric,
	}

	switch v := any(value).(type) {
	case Gauge:
		body.MType = models.Gauge
	case Counter:
		body.MType = models.Counter
	default:
		return fmt.Errorf("unsupported metric type %T", v)
	}

	err := a.Request(metric, endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to send: %v", err)
	}
	return nil
}

func (a Agent) Request(metric string, endpoint string, body models.Metrics) error {
	req := a.client.R().
		SetHeader("Content-Type", "application/json")

	switch a.config.EnableGzip {
	case true:
		buf, err := Compress(body)
		if err != nil {
			return fmt.Errorf("failed to compress request body: %v", err)
		}
		req.SetHeader("Content-Encoding", "gzip").
			SetHeader("Accept-Encoding", "gzip").
			SetBody(buf) // Use the compressed request body.
	case false:
		req.SetBody(body) // Use the uncompressed request body.
	}

	logger.Log.Debugf("Request body: ID = %s, MType = %s, Delta = %v, Value = %v", body.ID, body.MType, body.Delta, body.Value)
	logger.Log.Debug("Request header", req.Header)

	resp, err := req.Post(endpoint)
	if err != nil {
		return fmt.Errorf("failed to POST request: %v", err)
	}

	logger.Log.Debug("Response status‑code: ", resp.StatusCode(), " Metric: ", metric)
	logger.Log.Debug("Response header: ", resp.Header(), " Metric: ", metric)
	return nil
}

func Compress(data models.Metrics) ([]byte, error) {
	// Marshal the body to JSON.
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %v", err)
	}

	// Compress the JSON body.
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write(jsonBody)
	_ = zw.Close()

	return buf.Bytes(), nil
}
