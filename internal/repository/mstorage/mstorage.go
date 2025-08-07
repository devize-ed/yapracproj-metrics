package mstorage

import (
	"context"
	"fmt"
	"strconv"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
)

// MemStorage is the in-memory server storage for the metrics.
type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

// MemStorage constructor.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

// SetGauge sets the value of a gauge metric by its name.
func (ms *MemStorage) SetGauge(ctx context.Context, name string, value *float64) error {
	ms.Gauge[name] = *value
	return nil
}

// GetGauge retrieves the value of a gauge metric by its name.
func (ms *MemStorage) GetGauge(ctx context.Context, name string) (*float64, error) {
	val, ok := ms.Gauge[name]
	if !ok {
		return nil, fmt.Errorf("gauge %s not found", name)
	}
	return &val, nil
}

// AddCounter increments the value of a counter metric by the given delta.
func (ms *MemStorage) AddCounter(ctx context.Context, name string, delta *int64) error {
	ms.Counter[name] += *delta
	return nil

}

// GetCounter retrieves the value of a counter metric by its name.
func (ms *MemStorage) GetCounter(ctx context.Context, name string) (*int64, error) {
	val, ok := ms.Counter[name]
	if !ok {
		return nil, fmt.Errorf("counter %s not found", name)
	}
	return &val, nil
}

// SaveBatch saves a batch of metrics to the storage.
func (ms *MemStorage) SaveBatch(ctx context.Context, batch []models.Metrics) error {
	for _, m := range batch {
		switch m.MType {
		case models.Gauge:
			ms.Gauge[m.ID] = *m.Value
		case models.Counter:
			ms.Counter[m.ID] += *m.Delta
		}
	}
	return nil
}

// Get all the saved metrics from the storage and return them and values as strings.
func (ms *MemStorage) GetAll(ctx context.Context) (map[string]string, error) {
	result := make(map[string]string)
	for k, v := range ms.Gauge {
		result[k] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	for k, v := range ms.Counter {
		result[k] = strconv.FormatInt(v, 10)
	}
	return result, nil
}

// Ping is a no-op for the in-memory storage.
func (ms *MemStorage) Ping(ctx context.Context) error {
	return nil
}

// Close is a no-op for the in-memory storage.
func (ms *MemStorage) Close() error {
	return nil
}
