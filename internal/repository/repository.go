package repository

import (
	"strconv"
	"sync"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
)

// MemStorage is the in-memory server storage for the metrics
type MemStorage struct {
	mu       sync.RWMutex
	Gauge    map[string]float64 `json:"gauge"`
	Counter  map[string]int64   `json:"counter"`
	syncSave bool               `json:"-"`
	fpath    string             `json:"-"`
}

// MemStorage constructor
func NewMemStorage(storeInterval int, fpath string) *MemStorage {
	return &MemStorage{
		Gauge:    make(map[string]float64),
		Counter:  make(map[string]int64),
		syncSave: storeInterval == 0,
		fpath:    fpath,
	}
}

// Methods for acting the storage
func (ms *MemStorage) SetGauge(name string, value float64) {
	ms.mu.Lock()
	ms.Gauge[name] = value
	ms.mu.Unlock()

	if ms.syncSave {
		err := ms.Save(ms.fpath)
		if err != nil {
			logger.Log.Errorf("failed to save metrics: %v", err)
		}
	}
}

func (ms *MemStorage) GetGauge(name string) (float64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	val, ok := ms.Gauge[name]
	return val, ok
}

func (ms *MemStorage) AddCounter(name string, delta int64) {
	ms.mu.Lock()
	ms.Counter[name] += delta
	ms.mu.Unlock()

	if ms.syncSave {
		err := ms.Save(ms.fpath)
		if err != nil {
			logger.Log.Errorf("failed to save metrics: %v", err)
		}
	}
}

func (ms *MemStorage) GetCounter(name string) (int64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	val, ok := ms.Counter[name]
	return val, ok
}

// Get all the saved metrics from the storage and return them and values as strings
func (ms *MemStorage) ListAll() map[string]string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	result := make(map[string]string)
	for k, v := range ms.Gauge {
		result[k] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	for k, v := range ms.Counter {
		result[k] = strconv.FormatInt(v, 10)
	}
	return result
}
