package fstorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	cfg "github.com/devize-ed/yapracproj-metrics.git/internal/repository/fstorage/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository/mstorage"
)

// FileSaver is a struct that implements the Repository interface for saving metrics to a file.
type FileSaver struct {
	*mstorage.MemStorage
	fname    string
	syncSave bool
	mu       sync.RWMutex
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewFileSaver constructs a new FileSaver with the provided file name.
func NewFileSaver(ctx context.Context, config *cfg.FStorageConfig, storage *mstorage.MemStorage) *FileSaver {
	// Create the context that will be used to close the interval saver.
	fsCtx, cancel := context.WithCancel(ctx)

	// Initialize the FileSaver with the provided configuration and storage.
	fs := &FileSaver{
		MemStorage: storage, // internal storage to save the metrics to
		fname:      config.FPath,
		syncSave:   config.StoreInterval == 0, // if the interval is 0, saves synchronously
		cancel:     cancel,
	}

	// If the interval is not 0, start the interval saver.
	if config.StoreInterval != 0 {
		fs.wg.Add(1)
		go fs.intervalSaver(fsCtx, config.StoreInterval)
	}

	// If the restore is set, restore the metrics from the file.
	if config.Restore {
		if err := fs.restoreFromFile(ctx); err != nil {
			logger.Log.Errorf("failed to restore metrics from file: %v", err)
		}
	}

	return fs
}

// SetGauge sets the value of a gauge metric by its name.
func (f *FileSaver) SetGauge(ctx context.Context, name string, value *float64) error {
	f.mu.Lock()
	// Call the embedded MemStorage method
	if err := f.MemStorage.SetGauge(ctx, name, value); err != nil {
		f.mu.Unlock()
		return fmt.Errorf("failed to set gauge: %w", err)
	}

	if f.syncSave {
		// Save to file while holding the lock to ensure consistency
		if err := f.saveToFile(ctx); err != nil {
			f.mu.Unlock()
			return fmt.Errorf("failed to save metrics to file: %w", err)
		}
	}
	f.mu.Unlock()
	return nil
}

// GetGauge retrieves the value of a gauge metric by its name.
func (f *FileSaver) GetGauge(ctx context.Context, name string) (*float64, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Call the embedded MemStorage method
	val, err := f.MemStorage.GetGauge(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get gauge: %w", err)
	}

	return val, nil
}

// AddCounter increments the value of a counter metric by the given delta.
func (f *FileSaver) AddCounter(ctx context.Context, name string, delta *int64) error {
	f.mu.Lock()
	// Call the embedded MemStorage method
	if err := f.MemStorage.AddCounter(ctx, name, delta); err != nil {
		f.mu.Unlock()
		return fmt.Errorf("failed to add counter: %w", err)
	}

	// If the sync save is set, save the metrics to the file.
	if f.syncSave {
		if err := f.saveToFile(ctx); err != nil {
			f.mu.Unlock()
			return fmt.Errorf("failed to save metrics to file: %w", err)
		}
	}
	f.mu.Unlock()
	return nil
}

// GetCounter retrieves the value of a counter metric by its name.
func (f *FileSaver) GetCounter(ctx context.Context, name string) (*int64, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Call the embedded MemStorage method
	delta, err := f.MemStorage.GetCounter(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get counter: %w", err)
	}
	return delta, nil
}

// SaveBatch saves a batch of metrics to the repository.
func (f *FileSaver) SaveBatch(ctx context.Context, metrics []models.Metrics) error {
	logger.Log.Debugf("saving metrics to %s", f.fname)
	// Check if the file name is empty -> not saving (used in tests).
	if f.fname == "" {
		return nil
	}

	// Initialize maps to hold gauge and counter metrics.
	var gauge = map[string]float64{}
	var counter = map[string]int64{}

	// Iterate over the metrics and populate the gauge and counter maps.
	for _, m := range metrics {
		switch m.MType {
		case models.Gauge:
			gauge[m.ID] = *m.Value
		case models.Counter:
			counter[m.ID] = *m.Delta
		}
	}

	// Marshal the metrics to JSON format.
	data, err := json.MarshalIndent(struct {
		Gauge   map[string]float64 `json:"gauge"`
		Counter map[string]int64   `json:"counter"`
	}{gauge, counter}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// Check if the directory exists, create it if not.
	if err := os.MkdirAll(filepath.Dir(f.fname), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	// Write the data to the file.
	if err := writeFileWithRetries(ctx, f.fname, data); err != nil {
		return fmt.Errorf("failed to save metrics to file: %w", err)
	}

	f.mu.Lock()
	// Call the embedded MemStorage method
	if err := f.MemStorage.SaveBatch(ctx, metrics); err != nil {
		f.mu.Unlock()
		return fmt.Errorf("failed to save metrics to repository: %w", err)
	}
	f.mu.Unlock()

	logger.Log.Debugf("metrics saved (%d bytes) to %s", len(data), f.fname)
	return nil
}

func writeFileWithRetries(ctx context.Context, fname string, data []byte) error {
	backoffs := []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

	for attempt := 1; attempt <= len(backoffs)+1; attempt++ {
		if err := os.WriteFile(fname, data, 0o644); err != nil {
			if attempt == len(backoffs)+1 {
				return err
			}
			logger.Log.Debugf("write file failed, attempt %d: %v", attempt, err)
			select {
			case <-time.After(backoffs[attempt-1]):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}
	return nil
}

// Load reads the metrics from the specified file and restores them to the storage.
func (f *FileSaver) restoreFromFile(ctx context.Context) error {
	logger.Log.Debugf("loading metrics from %s", f.fname)
	// Check if the file name is empty -> not saving (used in tests).
	if f.fname == "" {
		return nil
	}

	// Check if the file exists.
	data, err := os.ReadFile(f.fname)
	if err != nil {
		return os.ErrNotExist
	}

	// Check if the data is empty.
	if len(data) == 0 {
		logger.Log.Warn("storage empty")
		return nil
	}

	// Unmarshal the data to a temporary struct.
	tmp := struct {
		Gauge   map[string]float64 `json:"gauge"`
		Counter map[string]int64   `json:"counter"`
	}{}

	if err := json.Unmarshal(data, &tmp); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			logger.Log.Warnf("data storage error: %v", err)
			return nil
		}
		return fmt.Errorf("error unmarshal metrics: %w", err)
	}

	f.mu.Lock()
	f.Gauge = tmp.Gauge
	f.Counter = tmp.Counter
	f.mu.Unlock()

	logger.Log.Debugf("metrics restored from %s", f.fname)
	return nil
}

func (f *FileSaver) saveToFile(ctx context.Context) error {
	logger.Log.Debugf("saving metrics to %s", f.fname)
	// Check if the file name is empty -> not saving (used in tests).
	if f.fname == "" {
		return nil
	}

	// Marshal the metrics to JSON format (lock is already held by caller)
	data, err := json.MarshalIndent(struct {
		Gauge   map[string]float64 `json:"gauge"`
		Counter map[string]int64   `json:"counter"`
	}{f.Gauge, f.Counter}, "", "  ")
	if err != nil {
		return err
	}

	// Check if the directory exists, create it if not.
	if err := os.MkdirAll(filepath.Dir(f.fname), 0o755); err != nil {
		return err
	}
	// Write the data to the file.
	if err := writeFileWithRetries(ctx, f.fname, data); err != nil {
		return err
	}
	logger.Log.Debugf("metrics saved (%d bytes) to %s", len(data), f.fname)
	return nil
}

func (f *FileSaver) intervalSaver(ctx context.Context, interval int) {
	defer f.wg.Done()

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := f.saveToFile(ctx); err != nil {
				logger.Log.Errorf("periodic save failed: %v", err)
			}
		case <-ctx.Done():
			logger.Log.Debug("Interval saver stopping, performing final save")
			return
		}
	}
}

// Close gracefully stops the ticker and saves the metrics to the file.
func (f *FileSaver) Close() error {
	// Cancel the context that is used to close the interval saver.
	if f.cancel != nil {
		f.cancel()
	}
	// Wait until the interval saver returns.
	f.wg.Wait()

	// Save the metrics to the file before the server is closed.
	if err := f.saveToFile(context.Background()); err != nil {
		return fmt.Errorf("final save failed: %w", err)
	}
	return nil
}
