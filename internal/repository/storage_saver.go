package repository

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"os"
	"path/filepath"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
)

// MemStorage is the in-memory server storage for the metrics.
func (ms *MemStorage) Save(fname string) error {
	logger.Log.Debugf("saving metrics to %s", fname)
	// Check if the file name is empty -> not saving (used in tests).
	if fname == "" {
		return nil
	}

	// Create a copy of the metrics to avoid holding the lock while marshaling.
	ms.mu.RLock()
	gCopy := maps.Clone(ms.Gauge)
	cCopy := maps.Clone(ms.Counter)
	ms.mu.RUnlock()

	// Marshal the metrics to JSON format.
	data, err := json.MarshalIndent(struct {
		Gauge   map[string]float64 `json:"gauge"`
		Counter map[string]int64   `json:"counter"`
	}{gCopy, cCopy}, "", "  ")
	if err != nil {
		return err
	}

	// Check if the directory exists, create it if not.
	if err := os.MkdirAll(filepath.Dir(fname), 0o755); err != nil {
		return err
	}
	// Write the data to the file.
	if err := os.WriteFile(fname, data, 0o644); err != nil {
		return err
	}
	logger.Log.Debugf("metrics saved (%d bytes) to %s", len(data), fname)
	return nil
}

// Load reads the metrics from the specified file and restores them to the storage.
func (ms *MemStorage) Load(fname string) error {
	logger.Log.Debugf("loading metrics from %s", fname)
	// Check if the file name is empty -> not saving (used in tests).
	if fname == "" {
		return nil
	}

	// Check if the file exists.
	data, err := os.ReadFile(fname)
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
		return err
	}

	// Update the storage with the loaded metrics.
	ms.mu.Lock()
	ms.Gauge, ms.Counter = tmp.Gauge, tmp.Counter
	ms.mu.Unlock()
	logger.Log.Debugf("metrics restored from %s", fname)
	return nil
}

// IntervalSaver periodically saves the metrics to the file
func (ms *MemStorage) IntervalSaver(ctx context.Context, interval int, fpath string) {
	// If the interval is 0, it saves only when the server is closing.
	logger.Log.Debugf("starting interval saver with interval %d seconds", interval)
	if interval == 0 {
		go func() {
			<-ctx.Done()
			if err := ms.Save(fpath); err != nil {
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
				if err := ms.Save(fpath); err != nil {
					logger.Log.Errorf("periodic save failed: %v", err)
				}
			case <-ctx.Done(): // Save the metrics before exiting
				if err := ms.Save(fpath); err != nil {
					logger.Log.Errorf("final save failed: %v", err)
				}
				return
			}
		}
	}()
}
