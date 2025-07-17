package agent

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"runtime"
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
			val := reflect.ValueOf(a.storage).Elem()
			typ := reflect.TypeOf(a.storage).Elem()

			for i := 0; i < val.NumField(); i++ {
				metric := typ.Field(i).Name
				value := val.Field(i)
				// fmt.Printf("%s = %v\n", metric, value)
				if err := SendMetric(a.client, metric, a.config.Host, value); err != nil {
					logger.Log.Error("metric ", metric, ": ", err)
				}
			}
		}
	}
}

// Sends a metric to the server
func SendMetric(client *resty.Client, metric, host string, value reflect.Value) error {
	logger.Log.Debug("SendMetric requested for metric: ", metric, " = ", value)

	body := models.Metrics{
		ID: metric,
	}

	switch metric {
	case "PollCount":
		if !isInteger(value.Kind()) {
			return fmt.Errorf("PollCount must be integer, got: %s", value.Kind())
		}
		body.MType = models.Counter
		v := value.Int()
		body.Delta = &v

	default:
		if !isFloat(value.Kind()) && !isInteger(value.Kind()) {
			return fmt.Errorf("unsupported kind for gauge: %s", value.Kind())
		}
		body.MType = models.Gauge

		var val float64
		if isInteger(value.Kind()) {
			val = float64(value.Int())
		} else {
			val = value.Float()
		}
		body.Value = &val
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

func isInteger(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Int64
}

func isFloat(k reflect.Kind) bool {
	return k == reflect.Float32 || k == reflect.Float64
}

// Metrics collector methods
func (m *AgentStorage) CollectMetrics() {
	logger.Log.Debug("Collecting metrics...")

	// read the metrics from the runtime package
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	// store metrics to the struct
	m.Alloc = stats.Alloc
	m.BuckHashSys = stats.BuckHashSys
	m.Frees = stats.Frees
	m.GCCPUFraction = stats.GCCPUFraction
	m.GCSys = stats.GCSys
	m.HeapAlloc = stats.HeapAlloc
	m.HeapIdle = stats.HeapIdle
	m.HeapInuse = stats.HeapInuse
	m.HeapObjects = stats.HeapObjects
	m.HeapReleased = stats.HeapReleased
	m.HeapSys = stats.HeapSys
	m.LastGC = stats.LastGC
	m.Lookups = stats.Lookups
	m.MCacheInuse = stats.MCacheInuse
	m.MCacheSys = stats.MCacheSys
	m.MSpanInuse = stats.MSpanInuse
	m.MSpanSys = stats.MSpanSys
	m.Mallocs = stats.Mallocs
	m.NextGC = stats.NextGC
	m.NumForcedGC = stats.NumForcedGC
	m.NumGC = stats.NumGC
	m.OtherSys = stats.OtherSys
	m.PauseTotalNs = stats.PauseTotalNs
	m.StackInuse = stats.StackInuse
	m.StackSys = stats.StackSys
	m.Sys = stats.Sys
	m.TotalAlloc = stats.TotalAlloc

	m.PollCount++                  // increment the poll count
	m.RandomValue = rand.Float64() // adda random value to the metrics

	logger.Log.Debug("All metrics collected")
}
