package agent

import (
	"fmt"
	"log"
	"math/rand/v2"
	"runtime"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/go-resty/resty/v2"
)

func SendMetric(client *resty.Client, metric, value, host string) error {

	log.Println("SendMetric requested for metric: ", metric, " = ", value)
	var mtype = models.Gauge
	if metric == "PollCount" {
		mtype = models.Counter
	}

	endpoint := fmt.Sprintf("http://%s/update/%s/%s/%s", host, mtype, metric, value)

	resp, err := client.R().
		SetHeader("Content-Type", "text/plain; charset=utf-8").
		Post(endpoint)
	if err != nil {
		log.Println("Error sending ", metric, ": ", err)
		return err
	}

	log.Println("Response status-code: ", resp.StatusCode())
	return nil
}

func (m *AgentStorage) CollectMetrics() {
	log.Println("Collecting metrics...")

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

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

	m.PollCount++
	m.RandomValue = rand.Float64()

	log.Println("All metrics collected")
}
