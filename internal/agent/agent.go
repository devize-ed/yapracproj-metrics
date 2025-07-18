package agent

import (
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
		storage: &AgentStorage{},
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
					logger.Log.Error("metric ", name, ": ", err)
				}
			}
			for name, val := range a.storage.Gauges {
				if err := SendMetric(a.client, name, a.config.Host, val); err != nil {
					logger.Log.Error("metric ", name, ": ", err)
				}
			}
		}
	}
}

// Sends a metric to the server
func SendMetric[T MetricValue](client *resty.Client, metric, host string, value T) error {

	body := models.Metrics{
		ID: metric,
	}

	switch v := any(value).(type) {
	case Gauge:
		body.MType = models.Counter
		floatValue := float64(v)
		body.Value = &floatValue
	case Counter:
		body.MType = models.Counter
		intValue := int64(v)
		body.Delta = &intValue
	default:
		return fmt.Errorf("unsupported metric type %T", v)
	}

	endpoint := fmt.Sprintf("http://%s/update", host)
	req := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body)
	logger.Log.Infoln("bosy for req: ", req.Body)
	resp, err := req.Post(endpoint)

	if err != nil {
		logger.Log.Error("Error sending ", metric, ": ", err)
		return err
	}

	logger.Log.Debug("Response status-code: ", resp.StatusCode(), " Metric: ", metric)
	return nil
}
