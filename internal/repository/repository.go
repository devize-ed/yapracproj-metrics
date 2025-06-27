package repository

import "strconv"

type Repository interface {
	SetGauge(name string, value float64)
	GetGauge(name string) (float64, bool)
	SetCounter(name string, value int64)
	GetCounter(name string) (int64, bool)
	ListAll()
}

// MemStorage is the in-memory server storage for the metrics
type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

// MemStorage constructor
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

// Methods for acting the storage
func (ms *MemStorage) SetGauge(name string, value float64) {
	ms.gauge[name] = value
}

func (ms *MemStorage) GetGauge(name string) (float64, bool) {
	val, ok := ms.gauge[name]
	return val, ok
}

func (ms *MemStorage) AddCounter(name string, delta int64) {
	ms.counter[name] += delta
}

func (ms *MemStorage) GetCounter(name string) (int64, bool) {
	val, ok := ms.counter[name]
	return val, ok
}

// Get all the saved metrics from the storage and return them and values as strings
func (ms *MemStorage) ListAll() map[string]string {
	result := make(map[string]string)
	for k, v := range ms.gauge {
		result[k] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	for k, v := range ms.counter {
		result[k] = strconv.FormatInt(v, 10)
	}
	return result
}
