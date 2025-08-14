package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/devize-ed/yapracproj-metrics.git/internal/sign"
	"github.com/go-resty/resty/v2"
)

// Agent holds the HTTP client, metric storage, and configuration.
type Agent struct {
	client    *resty.Client
	storage   *AgentStorage
	config    config.AgentConfig
	wg        sync.WaitGroup
	jobsQueue chan batchRequest
}

type batchRequest struct {
	name      string
	endpoint  string
	bodyBytes []byte
}

// NewAgent returns a new Agent that uses the given HTTP client and configuration.
func NewAgent(client *resty.Client, config config.AgentConfig) *Agent {
	return &Agent{
		client:  clientWithRetries(client),
		storage: NewAgentStorage(),
		config:  config,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	// Get the number of workers.
	numWorkers := a.config.Agent.RateLimit
	// Create channels.
	a.jobsQueue = make(chan batchRequest, 2*numWorkers)
	defer close(a.jobsQueue)

	// Create a worker pool.
	errCh := a.createWorkerPool(numWorkers)

	// Convert interval values to time.Duration.
	timePollInterval := time.Duration(a.config.Agent.PollInterval) * time.Second
	timeReportInterval := time.Duration(a.config.Agent.ReportInterval) * time.Second

	// Set up the tickers for polling and reporting.
	pollTicker := time.NewTicker(timePollInterval)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(timeReportInterval)
	defer reportTicker.Stop()

	// Start the agent loop.
	for {
		select {
		case <-pollTicker.C: // Collect metrics at the polling interval.
			a.storage.CollectMetrics(ctx)
		case <-reportTicker.C: // Send metrics at the reporting interval.
			// Check whether "test‑get" mode is enabled.
			if !a.config.Agent.EnableTestGet {
				// Send metrics as a batch to the server.
				metrics := a.LoadMetrics()
				if err := a.SendMetricsBatch(metrics); err != nil {
					logger.Log.Error("error sending batch metrics: ", err)
				}
			} else {
				logger.Log.Debug("Test‑get mode enabled, skipping sending metrics.")
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
		case err := <-errCh:
			return fmt.Errorf("error running agent: %v", err)
		case <-ctx.Done():
			logger.Log.Debug("Closing agent")
			// Close the error channel.
			close(errCh)
			// Wait for the workers to finish.
			a.wg.Wait()
			return nil
		}
	}
}

// LoadMetrics loads metrics from the agent storage.
func (a *Agent) LoadMetrics() []models.Metrics {
	a.storage.mu.RLock()
	defer a.storage.mu.RUnlock()

	var metrics = []models.Metrics{}
	// Load gauges.
	for name, val := range a.storage.Gauges {
		floatVal := float64(val)
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &floatVal,
		})
	}
	// Load counters.
	for name, val := range a.storage.Counters {
		intVal := int64(val)
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &intVal,
		})
	}
	return metrics
}

// SendMetricsBatch sends a batch of metrics to the server.
func (a *Agent) SendMetricsBatch(metrics []models.Metrics) error {
	logger.Log.Debug("Reporting metrics...")

	endpoint := fmt.Sprintf("http://%s/updates/", a.config.Connection.Host)

	// Marshal the body to JSON.
	bodyBytes, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("error marshalling request body: %v", err)
	}

	// if len(a.jobsQueue) == cap(a.jobsQueue) {
	// 	<-a.jobsQueue
	// }
	a.jobsQueue <- batchRequest{
		name:      "batch",
		endpoint:  endpoint,
		bodyBytes: bodyBytes,
	}
	return nil
}

// GetMetric requests a metric from the server for testing purposes.
func GetMetric[T MetricValue](a *Agent, metric string, value T) error {
	endpoint := fmt.Sprintf("http://%s/value/", a.config.Connection.Host)

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

	// Marshal the body to JSON.
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshalling request body: %v", err)
	}

	err = a.Request(metric, endpoint, bodyBytes)
	if err != nil {
		return fmt.Errorf("failed to send: %v", err)
	}
	return nil
}

func (a *Agent) Request(name string, endpoint string, bodyBytes []byte) error {
	logger.Log.Debugf("Request: %s %s", name, endpoint)

	// Create a new request.
	req := a.client.R().
		SetHeader("Content-Type", "application/json")

	var body []byte
	// Compress the request body if the gzip is enabled.
	if a.config.Agent.EnableGzip {
		req.SetHeader("Content-Encoding", "gzip").
			SetHeader("Accept-Encoding", "gzip")
		body = Compress(bodyBytes)
	} else {
		body = bodyBytes
	}

	// Set the hash of the request body.
	if a.config.Sign.Key != "" {
		logger.Log.Debugf("Setting hash header")
		hash := sign.Hash(body, a.config.Sign.Key)
		req.SetHeader(sign.HashHeader, hash)
	}

	// Set the request body.
	req.SetBody(body)

	logger.Log.Debugf("Request body: %s", string(bodyBytes))
	logger.Log.Debugf("Request header: %v", req.Header)

	resp, err := req.Post(endpoint)
	if err != nil {
		return fmt.Errorf("failed to POST request: %v", err)
	}

	logger.Log.Debugf("Response status-code: %d", resp.StatusCode())
	logger.Log.Debugf("Response header: %v", resp.Header())
	return nil
}

func Compress(data []byte) []byte {
	logger.Log.Debugf("Compressing data")
	// Compress the JSON body.
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write(data)
	_ = zw.Close()

	return buf.Bytes()
}

func clientWithRetries(client *resty.Client) *resty.Client {
	// Set the retry count and backoff delay.
	backoffs := []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

	// Set the retry count and backoff delay.
	client.SetRetryCount(len(backoffs)).
		SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
			// Get the retry count.
			n := r.Request.Attempt - 1
			if n >= len(backoffs) {
				n = len(backoffs) - 1
			}
			// Get the backoff delay.
			delay := backoffs[n]
			logger.Log.Debugf("retry attempt %d, waiting %s", r.Request.Attempt, delay)
			return delay, nil
		}).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			// Check if the error is retryable.
			if err != nil && isErrorRetryable(err) {
				logger.Log.Warnf("network error: %v — will retry", err)
				return true
			}

			return false
		})
	return client
}

// isErrorRetryable checks if the error is retryable.
func isErrorRetryable(err error) bool {
	// Check if the error is a network error.
	var ne net.Error
	return errors.As(err, &ne)
}
