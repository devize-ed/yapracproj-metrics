package agent

import (
	"math/rand/v2"
	"runtime"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
)

type (
	Gauge   float64
	Counter int64
)

// generics for Metric value
type MetricValue interface {
	Gauge | Counter
}

// struct to hold the metrics collected by the agent
type AgentStorage struct {
	Counters map[string]Counter
	Gauges   map[string]Gauge
}

func NewAgentStorage() *AgentStorage {
	return &AgentStorage{
		Counters: make(map[string]Counter),
		Gauges:   make(map[string]Gauge),
	}
}

// Metrics collector methods
func (s *AgentStorage) CollectMetrics() {
	logger.Log.Debug("Collecting metrics")

	// read the metrics from the runtime package
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// store metrics to the storage
	s.Gauges["Alloc"] = Gauge(m.Alloc)
	s.Gauges["BuckHashSys"] = Gauge(m.BuckHashSys)
	s.Gauges["Frees"] = Gauge(m.Frees)
	s.Gauges["GCCPUFraction"] = Gauge(m.GCCPUFraction)
	s.Gauges["GCSys"] = Gauge(m.GCSys)
	s.Gauges["HeapAlloc"] = Gauge(m.HeapAlloc)
	s.Gauges["HeapIdle"] = Gauge(m.HeapIdle)
	s.Gauges["HeapInuse"] = Gauge(m.HeapInuse)
	s.Gauges["HeapObjects"] = Gauge(m.HeapObjects)
	s.Gauges["HeapReleased"] = Gauge(m.HeapReleased)
	s.Gauges["HeapSys"] = Gauge(m.HeapSys)
	s.Gauges["LastGC"] = Gauge(m.LastGC)
	s.Gauges["Lookups"] = Gauge(m.Lookups)
	s.Gauges["MCacheInuse"] = Gauge(m.MCacheInuse)
	s.Gauges["MCacheSys"] = Gauge(m.MCacheSys)
	s.Gauges["MSpanInuse"] = Gauge(m.MSpanInuse)
	s.Gauges["MSpanSys"] = Gauge(m.MSpanSys)
	s.Gauges["Mallocs"] = Gauge(m.Mallocs)
	s.Gauges["NextGC"] = Gauge(m.NextGC)
	s.Gauges["NumForcedGC"] = Gauge(m.NumForcedGC)
	s.Gauges["NumGC"] = Gauge(m.NumGC)
	s.Gauges["OtherSys"] = Gauge(m.OtherSys)
	s.Gauges["PauseTotalNs"] = Gauge(m.PauseTotalNs)
	s.Gauges["StackInuse"] = Gauge(m.StackInuse)
	s.Gauges["StackSys"] = Gauge(m.StackSys)
	s.Gauges["Sys"] = Gauge(m.Sys)
	s.Gauges["TotalAlloc"] = Gauge(m.TotalAlloc)

	// additional values
	s.Counters["PollCount"]++                       // increment the poll count
	s.Gauges["RandomValue"] = Gauge(rand.Float64()) // add a random value to the metrics
}
