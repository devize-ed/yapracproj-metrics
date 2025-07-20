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

// save current metric values to the file
func (ms *MemStorage) Save(fname string) error {
	logger.Log.Debugf("saving metrics to %s", fname)
	// check if the file name is empty -> not saving (used in tests)
	if fname == "" {
		return nil
	}

	// create a copy of the metrics to avoid holding the lock while marshaling
	ms.mu.RLock()
	gCopy := maps.Clone(ms.Gauge)
	cCopy := maps.Clone(ms.Counter)
	ms.mu.RUnlock()

	// marshal the metrics to JSON format
	data, err := json.MarshalIndent(struct {
		Gauge   map[string]float64 `json:"gauge"`
		Counter map[string]int64   `json:"counter"`
	}{gCopy, cCopy}, "", "  ")
	if err != nil {
		return err
	}

	// check if the directory exists, create it if not
	if err := os.MkdirAll(filepath.Dir(fname), 0o755); err != nil {
		return err
	}
	// write the data to the file
	if err := os.WriteFile(fname, data, 0o644); err != nil {
		return err
	}
	logger.Log.Debugf("metrics saved (%d bytes) to %s", len(data), fname)
	return nil
}

// load metric values from the file
func (ms *MemStorage) Load(fname string) error {
	logger.Log.Debugf("loading metrics from %s", fname)
	// check if the file name is empty -> not saving (used in tests)
	if fname == "" {
		return nil
	}

	// check if the file exists
	data, err := os.ReadFile(fname)
	if err != nil {
		return os.ErrNotExist
	}

	// check if the data is empty
	if len(data) == 0 {
		logger.Log.Warn("storage empty")
		return nil
	}

	// unmarshal the data to a temporary struct
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

	// update the storage with the loaded metrics
	ms.mu.Lock()
	ms.Gauge, ms.Counter = tmp.Gauge, tmp.Counter
	ms.mu.Unlock()
	logger.Log.Debugf("metrics restored from %s", fname)
	return nil
}

// IntervalSaver periodically saves the metrics to the file
// if the interval is 0, it saves only the server is closing
func (ms *MemStorage) IntervalSaver(ctx context.Context, interval int, fpath string) {
	// if the interval is 0, it saves only when the server is closing
	if interval == 0 {
		go func() {
			<-ctx.Done()
			if err := ms.Save(fpath); err != nil {
				logger.Log.Errorf("final save (sync mode) failed: %v", err)
			}
		}()
		return
	}

	// create a ticker to save the metrics on interval
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		// start ticker loop
		for {
			select {
			case <-ticker.C: // save the metrics on  interval
				if err := ms.Save(fpath); err != nil {
					logger.Log.Errorf("periodic save failed: %v", err)
				}
			case <-ctx.Done(): // save the metrics before exiting
				if err := ms.Save(fpath); err != nil {
					logger.Log.Errorf("final save failed: %v", err)
				}
				return
			}
		}
	}()
}
