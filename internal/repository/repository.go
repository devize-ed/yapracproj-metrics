package repository

import (
	"context"
	"fmt"
	"maps"
	"strconv"
	"sync"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
)

// Repository interface defines the methods for saving and loading metrics.
type Repository interface {
	Save(ctx context.Context, gauge map[string]float64, counter map[string]int64) error
	Load(ctx context.Context) (map[string]float64, map[string]int64, error)
}

// MemStorage is the in-memory server storage for the metrics.
type MemStorage struct {
	mu         sync.RWMutex
	Gauge      map[string]float64 `json:"gauge"`
	Counter    map[string]int64   `json:"counter"`
	syncSave   bool               `json:"-"`
	repository Repository         `json:"-"`
}

// MemStorage constructor.
func NewMemStorage(storeInterval int, r Repository) *MemStorage {
	return &MemStorage{
		Gauge:      make(map[string]float64),
		Counter:    make(map[string]int64),
		syncSave:   storeInterval == 0,
		repository: r,
	}
}

// SetGauge sets the value of a gauge metric by its name.
func (ms *MemStorage) SetGauge(ctx context.Context, name string, value float64) {
	ms.mu.Lock()
	ms.Gauge[name] = value
	ms.mu.Unlock()

	if ms.syncSave {
		err := ms.SaveToRepo(ctx)
		if err != nil {
			logger.Log.Errorf("failed to save metrics: %v", err)
		}
	}
}

// GetGauge retrieves the value of a gauge metric by its name.
func (ms *MemStorage) GetGauge(ctx context.Context, name string) (float64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	val, ok := ms.Gauge[name]
	return val, ok
}

// AddCounter increments the value of a counter metric by the given delta.
func (ms *MemStorage) AddCounter(ctx context.Context, name string, delta int64) {
	ms.mu.Lock()
	ms.Counter[name] += delta
	ms.mu.Unlock()

	if ms.syncSave {
		err := ms.SaveToRepo(ctx)
		if err != nil {
			logger.Log.Errorf("failed to save metrics: %v", err)
		}
	}
}

// GetCounter retrieves the value of a counter metric by its name.
func (ms *MemStorage) GetCounter(ctx context.Context, name string) (int64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	val, ok := ms.Counter[name]
	return val, ok
}

// Get all the saved metrics from the storage and return them and values as strings.
func (ms *MemStorage) ListAll(ctx context.Context) map[string]string {
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

// SaveToRepo writes the metrics to the repository.
func (ms *MemStorage) SaveToRepo(ctx context.Context) error {

	// copy the metrics to avoid holding the lock while saving.
	ms.mu.RLock()
	gCopy := maps.Clone(ms.Gauge)
	cCopy := maps.Clone(ms.Counter)
	ms.mu.RUnlock()

	// Save the metrics to the repository.
	if err := ms.repository.Save(ctx, gCopy, cCopy); err != nil {
		return fmt.Errorf("failed to save metrics: %w", err)
	}
	return nil

}

// LoadFromRepo retrieves the metrics from the repository and restores them to the storage.
func (ms *MemStorage) LoadFromRepo(ctx context.Context) error {
	// Load the metrics from the repository.
	gauge, counter, err := ms.repository.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load metrics: %w", err)
	}

	// Update the storage with the loaded metrics.
	ms.mu.Lock()
	ms.Gauge = gauge
	ms.Counter = counter
	ms.mu.Unlock()

	return nil
}

// IntervalSaver periodically saves the metrics to the file
func (ms *MemStorage) IntervalSaver(ctx context.Context, interval int) {
	// If the interval is 0, it saves only when the server is closing.
	logger.Log.Debugf("starting interval saver with interval %d seconds", interval)
	if interval == 0 {
		go func() {
			<-ctx.Done()
			if err := ms.SaveToRepo(ctx); err != nil {
				logger.Log.Errorf("final save (sync mode) failed: %v", err)
			}
		}()
		return
	}

	// Create a ticker to save the metrics on interval
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		// Start ticker loop.
		for {
			select {
			case <-ticker.C: // Save the metrics on interval
				if err := ms.SaveToRepo(ctx); err != nil {
					logger.Log.Errorf("periodic save failed: %v", err)
				}
			case <-ctx.Done(): // Save the metrics before exiting
				if err := ms.SaveToRepo(ctx); err != nil {
					logger.Log.Errorf("final save failed: %v", err)
				}
				return
			}
		}
	}()
}
