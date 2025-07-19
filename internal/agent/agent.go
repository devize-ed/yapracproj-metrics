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

// holds the client, storage and configuration for the agent
type Agent struct {
	client  *resty.Client
	storage *AgentStorage
	config  config.AgentConfig
}

// initializes a new Agent instance with the provided client and cfg
func NewAgent(client *resty.Client, config config.AgentConfig) *Agent {
	return &Agent{
		client:  client,
		storage: NewAgentStorage(),
		config:  config,
	}
}

func (a *Agent) Run() error {
	// convert the interval values to time.Duration
	timePollInterval := time.Duration(a.config.PollInterval) * time.Second
	timeReportInterval := time.Duration(a.config.ReportInterval) * time.Second

	// set up the ticker for polling and reporting
	pollTicker := time.NewTicker(timePollInterval)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(timeReportInterval)
	defer reportTicker.Stop()

	// start agent loop
	for {
		select {
		case <-pollTicker.C: // collect metrics at the polling interval
			a.storage.CollectMetrics()
		case <-reportTicker.C: // send metrics at the reporting interval
			logger.Log.Debug("Reporting metrics...")

			// iterate over the agent storage and send metrics to the server
			for name, val := range a.storage.Counters {
				if err := SendMetric(a.client, name, a.config.Host, val); err != nil {
					logger.Log.Error("error sending ", name, ": ", err)
				}
			}
			for name, val := range a.storage.Gauges {
				if err := SendMetric(a.client, name, a.config.Host, val); err != nil {
					logger.Log.Error("error sending ", name, ": ", err)
				}
			}

			// for name, val := range a.storage.Counters {
			// 	if err := GetMetric(a.client, name, a.config.Host, val); err != nil {
			// 		logger.Log.Error("error getting ", name, ": ", err)
			// 	}
			// }
			// for name, val := range a.storage.Gauges {
			// 	if err := GetMetric(a.client, name, a.config.Host, val); err != nil {
			// 		logger.Log.Error("error getting ", name, ": ", err)
			// 	}
			// }
		}
	}
}

// sends a metric to the server
func SendMetric[T MetricValue](client *resty.Client, metric, host string, value T) error {

	body := models.Metrics{
		ID: metric,
	}

	// set value and metric type
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

	// marshl bosy to bytes
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshling request body: %v", err)
	}

	// compress the body
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write(jsonBody)
	_ = zw.Close()

	// make and send request
	endpoint := fmt.Sprintf("http://%s/update/", host)
	req := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(buf.Bytes()) // set compressed body for the request

	logger.Log.Debugf("req body: ID = %s, MType = %s, Delta = %v, Value = %v", body.ID, body.MType, body.Delta, body.Value)
	resp, err := req.Post(endpoint)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	logger.Log.Debug("Response status-code: ", resp.StatusCode(), " Metric: ", metric)
	logger.Log.Debug("Response header: ", resp.Header(), " Metric: ", metric)
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		logger.Log.Fatal(err)
	}
	fmt.Printf("%+v\n", body)

	return nil
}

// test function for testing server
func GetMetric[T MetricValue](client *resty.Client, metric, host string, value T) error {
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

	// marshl body to bytes
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshling request body: %v", err)
	}

	// compress the body
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write(jsonBody)
	_ = zw.Close()

	// make and send request
	endpoint := fmt.Sprintf("http://%s/value/", host)
	req := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(buf.Bytes()) // set compressed body for the request
		// SetBody(body)

	logger.Log.Debugf("req body: ID = %s, MType = %s, Delta = %v, Value = %v", body.ID, body.MType, body.Delta, body.Value)
	logger.Log.Debug("req header", req.Header)
	resp, err := req.Post(endpoint)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	logger.Log.Debug("Response status-code: ", resp.StatusCode(), " Metric: ", metric)
	logger.Log.Debug("Response header: ", resp.Header(), " Metric: ", metric)

	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		logger.Log.Fatal(err)
	}
	fmt.Printf("%+v\n", body)
	return nil

}
