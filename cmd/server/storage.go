package main

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

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
