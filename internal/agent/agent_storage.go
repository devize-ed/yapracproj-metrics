package agent

import (
	"math/rand/v2"
	"runtime"
	"sync"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/shirou/gopsutil/v4/mem"
)

// Metric types for the agent storage
type (
	Gauge   float64
	Counter int64
)

// MetricValue is a type constraint for metric values.
type MetricValue interface {
	Gauge | Counter
}

// AgentStorage holds the metrics collected by the agent.
type AgentStorage struct {
	mu       sync.Mutex
	Counters map[string]Counter
	Gauges   map[string]Gauge
}

// NewAgentStorage initializes a new AgentStorage instance with empty maps for counters and gauges.
func NewAgentStorage() *AgentStorage {
	return &AgentStorage{
		Counters: make(map[string]Counter),
		Gauges:   make(map[string]Gauge),
	}
}

// CollectMetrics collects and store metrics.
func (s *AgentStorage) CollectMetrics() {
	logger.Log.Debug("Collecting metrics")
	s.collectRuntimeMetrics()
	s.collectAdditionalMetrics()
	s.collectSystemMetrics()
}

// collectRuntimeMetrics collects runtime metrics and stores them in the agent storage.
func (s *AgentStorage) collectRuntimeMetrics() {
	// Read metrics from the runtime package.
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	s.mu.Lock()
	// Store metrics to storage.
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
	s.mu.Unlock()
}

// collectAdditionalMetrics adds additional metrics to the agent storage.
func (s *AgentStorage) collectAdditionalMetrics() {
	s.mu.Lock()
	s.Counters["PollCount"]++                       // Increment the poll count
	s.Gauges["RandomValue"] = Gauge(rand.Float64()) // Add a random value to the metrics.
	s.mu.Unlock()
}

// collectSystemMetrics collects system metrics and stores them in the agent storage.
func (s *AgentStorage) collectSystemMetrics() {
	m, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Error("Error collecting system metrics: ", err)
	}
	s.mu.Lock()
	s.Gauges["TotalMemory"] = Gauge(m.Total)
	s.Gauges["FreeMemory"] = Gauge(m.Free)
	s.Gauges["CPUutilization1"] = Gauge(m.UsedPercent)
	s.mu.Unlock()
}
