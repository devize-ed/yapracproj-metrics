package agent

import (
	"fmt"
	"log"
	"math/rand/v2"
	"runtime"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/go-resty/resty/v2"
)

// SendMetric sends a metric to the server
func SendMetric(client *resty.Client, metric, value, host string) error {

	log.Println("SendMetric requested for metric: ", metric, " = ", value)

	// set the metric type as gauge by default, change it for counter if metric == "PollCount"
	var mtype = models.Gauge
	if metric == "PollCount" {
		mtype = models.Counter
	}

	// build the request URL and call the POST method request to the server	//
	endpoint := fmt.Sprintf("http://%s/update/%s/%s/%s", host, mtype, metric, value)
	resp, err := client.R().
		SetHeader("Content-Type", "text/plain; charset=utf-8").
		Post(endpoint)
	if err != nil {
		log.Println("Error sending ", metric, ": ", err)
		return err
	}

	// logging the response status code
	log.Println("Response status-code: ", resp.StatusCode())
	return nil
}

// Metrics collector methods
func (m *AgentStorage) CollectMetrics() {
	log.Println("Collecting metrics...")

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

	log.Println("All metrics collected")
}
