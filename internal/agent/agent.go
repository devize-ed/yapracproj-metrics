// Package agent provides functionality for collecting and sending metrics to a server.
// It includes metric collection, batching, compression, and retry mechanisms.
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
	"github.com/devize-ed/yapracproj-metrics.git/internal/encryption"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/devize-ed/yapracproj-metrics.git/internal/sign"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const batchSize = 10

// Agent holds the HTTP client, metric storage, and configuration.
type Agent struct {
	client  *resty.Client
	storage *AgentStorage
	config  config.AgentConfig
	logger  *zap.SugaredLogger
}

type jobs struct {
	wg        sync.WaitGroup
	jobsQueue chan batchRequest
	logger    *zap.SugaredLogger
}

type batchRequest struct {
	name      string
	endpoint  string
	bodyBytes []byte
}

// NewAgent returns a new Agent that uses the given HTTP client and configuration.
func NewAgent(client *resty.Client, config config.AgentConfig, logger *zap.SugaredLogger) *Agent {
	return &Agent{
		client:  clientWithRetries(client, logger),
		storage: NewAgentStorage(logger),
		config:  config,
		logger:  logger,
	}
}

// NewJobs creates a new jobs queue with the specified number of workers.
func NewJobs(numWorkers int, logger *zap.SugaredLogger) *jobs {
	return &jobs{
		jobsQueue: make(chan batchRequest, numWorkers),
		logger:    logger,
	}
}

// Run starts the agent's main loop for collecting and sending metrics.
func (a *Agent) Run(ctx context.Context) error {
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
			a.gatherMetrics()
		case <-reportTicker.C: // Send metrics at the reporting interval.
			if err := a.sendMetrics(); err != nil {
				a.logger.Error("error reporting metrics: %w", err)
			}
		case <-ctx.Done():
			a.logger.Info("Stop signal received, sending remaining metrics...")
			// Perform a final collection to capture the latest values.
			a.gatherMetrics()
			// Bound network operations during shutdown.
			a.client.SetTimeout(time.Duration(a.config.ShutdownTimeout))
			if err := a.sendMetrics(); err != nil {
				a.logger.Errorf("final send failed: %v", err)
			}
			a.logger.Debug("Closing agent")
			return nil
		}
	}
}

// gatherMetrics collects metrics from the system.
func (a *Agent) gatherMetrics() {
	a.storage.collectMetrics()
}

// sendMetrics sends collected metrics to the server.
func (a *Agent) sendMetrics() error {
	a.logger.Debug("Sending metrics")
	// Check whether "test‑get" mode is enabled.
	if !a.config.Agent.EnableTestGet {
		// Get the number of workers.
		numWorkers := a.config.Agent.RateLimit
		// Create a jobs queue.
		a.logger.Debug("Creating jobs queue with ", numWorkers, " workers")
		jobs := NewJobs(numWorkers, a.logger)
		// Create a worker pool.
		errCh := jobs.createWorkerPool(a.request, numWorkers, a.logger)

		// Send metrics as a batch to the server.
		metrics := a.loadMetrics()
		if err := jobs.sendMetricsBatch(a.config.Connection.Host, metrics, a.logger); err != nil {
			return fmt.Errorf("error sending batch metrics: %w", err)
		}
		// Close the jobs queue after pushing all metrics.
		close(jobs.jobsQueue)

		// Wait for the workers to finish.
		jobs.wg.Wait()
		// Close the error channel.
		close(errCh)
		// Check for errors.
		var errs error
		for err := range errCh {
			errs = errors.Join(errs, err)
		}
		return errs
	} else {
		a.logger.Debug("Test‑get mode enabled, skipping sending metrics.")
		// “Test‑get” mode: request metrics from the server.
		// Copy the metrics from the storage to the temporary maps.
		a.storage.mu.RLock()
		tmpGauges := make(map[string]Gauge, len(a.storage.Gauges))
		for metric, value := range a.storage.Gauges {
			tmpGauges[metric] = value
		}
		tmpCounters := make(map[string]Counter, len(a.storage.Counters))
		for metric, delta := range a.storage.Counters {
			tmpCounters[metric] = delta
		}
		a.storage.mu.RUnlock()
		// Get the metrics from the server.
		for name, val := range tmpCounters {
			if err := getMetric(a.request, a.config.Connection.Host, name, val); err != nil {
				return fmt.Errorf("error getting %s: %w", name, err)
			}
		}
		for name, val := range tmpGauges {
			if err := getMetric(a.request, a.config.Connection.Host, name, val); err != nil {
				return fmt.Errorf("error getting %s: %w", name, err)
			}
		}
	}
	return nil
}

// LoadMetrics loads metrics from the agent storage.
func (a *Agent) loadMetrics() []models.Metrics {
	a.logger.Debug("Loading metrics snapshot")
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
func (j *jobs) sendMetricsBatch(host string, metrics []models.Metrics, logger *zap.SugaredLogger) error {
	// Нужно разделить metrics на батчи по N метрик (const) и отправить в jobsQueue
	for i := 0; i < len(metrics); i += batchSize {
		batch := metrics[i:min(i+batchSize, len(metrics))]

		logger.Debugf("Sending metrics batch to %s", host)

		endpoint := fmt.Sprintf("http://%s/updates/", host)

		// Marshal the body to JSON.
		bodyBytes, err := json.Marshal(batch)
		if err != nil {
			return fmt.Errorf("error marshalling request body: %w", err)
		}

		j.jobsQueue <- batchRequest{
			name:      fmt.Sprintf("batch %d", i/batchSize),
			endpoint:  endpoint,
			bodyBytes: bodyBytes,
		}
	}
	logger.Debugf("All batches sent")
	return nil
}

// GetMetric requests a metric from the server for testing purposes.
func getMetric[T MetricValue](request func(name string, endpoint string, bodyBytes []byte) error, host, metric string, value T) error {
	endpoint := fmt.Sprintf("http://%s/value/", host)

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
		return fmt.Errorf("error marshalling request body: %w", err)
	}

	err = request(metric, endpoint, bodyBytes)
	if err != nil {
		return fmt.Errorf("failed to send: %w", err)
	}
	return nil
}

// request sends an HTTP request to the specified endpoint.
func (a *Agent) request(name string, endpoint string, bodyBytes []byte) error {
	a.logger.Debugf("Request: %s %s", name, endpoint)

	// Create a new request.
	req := a.client.R().
		SetHeader("Content-Type", "application/json")

	var (
		body []byte
		err  error
	)
	// Compress the request body if the gzip is enabled.
	if a.config.Agent.EnableGzip {
		req.SetHeader("Content-Encoding", "gzip").
			SetHeader("Accept-Encoding", "gzip")
		body, err = compress(bodyBytes, a.logger)
		if err != nil {
			return fmt.Errorf("failed to compress request body: %w", err)
		}
	} else {
		body = bodyBytes
	}

	// Encrypt the request body if the encryption is enabled.
	if a.config.Encryption.CryptoKey != "" {
		encryptor, err := encryption.NewEncryptor(a.config.Encryption.CryptoKey)
		if err != nil {
			return fmt.Errorf("failed to create encryptor: %w", err)
		}
		body, err = encryptor.Encrypt(body)
		if err != nil {
			return fmt.Errorf("failed to encrypt request body: %w", err)
		}
		req.SetHeader("Content-Type", "application/octet-stream").SetHeader("X-Encryption", "rsa")
	}

	// Set the hash of the request body.
	if a.config.Sign.Key != "" {
		a.logger.Debugf("Setting hash header")
		hash := sign.Hash(body, a.config.Sign.Key)
		req.SetHeader(sign.HashHeader, hash)
	}

	// Set the request body.
	req.SetBody(body)

	a.logger.Debugf("Request body: %s", string(bodyBytes))
	a.logger.Debugf("Request header: %v", req.Header)

	resp, err := req.Post(endpoint)
	if err != nil {
		return fmt.Errorf("failed to POST request: %w", err)
	}

	a.logger.Debugf("Response status-code: %d", resp.StatusCode())
	a.logger.Debugf("Response header: %v", resp.Header())
	return nil
}

// compress compresses data using gzip compression.
func compress(data []byte, logger *zap.SugaredLogger) ([]byte, error) {
	logger.Debugf("Compressing data")
	// Compress the JSON body.
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		return nil, fmt.Errorf("gzip write failed: %w", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("gzip close failed: %w", err)
	}

	return buf.Bytes(), nil
}

// clientWithRetries configures the HTTP client with retry logic and backoff.
func clientWithRetries(client *resty.Client, logger *zap.SugaredLogger) *resty.Client {
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
			logger.Debugf("retry attempt %d, waiting %s", r.Request.Attempt, delay)
			return delay, nil
		}).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			// Check if the error is retryable.
			if err != nil && isErrorRetryable(err) {
				logger.Warnf("network error: %w — will retry", err)
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
