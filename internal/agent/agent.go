package agent

import (
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"runtime"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
)

func SendMetric(client *http.Client, metric, value, host string) error {
	fmt.Println("SendMetric requested for metric: ", metric, " = ", value)
	var mtype = models.Gauge
	if metric == "PollCount" {
		mtype = models.Counter
	}

	endpoint := fmt.Sprintf("http://%s/update/%s/%s/%s", host, mtype, metric, value)
	request, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		fmt.Println("Error creating POST request for ", metric, ": ", err)
		return err
	}

	request.Header.Add("Content-Type", "text/plain; charset=utf-8")

	fmt.Println("Sending metric to ", endpoint)
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error sending ", metric, ": ", err)
		return err
	}

	fmt.Println("Response status-code: ", response.Status)

	defer response.Body.Close()
	_, err = io.Copy(io.Discard, response.Body)
	if err != nil {
		fmt.Println("Error discarding response body: ", err)
		return err
	}
	return nil
}

func (m *AgentStorage) CollectMetrics() {
	fmt.Println("Collecting metrics...")

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

	fmt.Println("All metrics collected")
}
