package fsaver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
)

// FileSaver is a struct that implements the Repository interface for saving metrics to a file.
type FileSaver struct {
	fname string // File name to save the metrics.
}

// NewFileSaver constructs a new FileSaver with the provided file name.
func NewFileSaver(fname string) *FileSaver {
	return &FileSaver{
		fname: fname,
	}
}

// Save writes the metrics to the specified file in JSON format.
func (f *FileSaver) Save(ctx context.Context, gauge map[string]float64, counter map[string]int64) error {
	logger.Log.Debugf("saving metrics to %s", f.fname)
	// Check if the file name is empty -> not saving (used in tests).
	if f.fname == "" {
		return nil
	}

	// Marshal the metrics to JSON format.
	data, err := json.MarshalIndent(struct {
		Gauge   map[string]float64 `json:"gauge"`
		Counter map[string]int64   `json:"counter"`
	}{gauge, counter}, "", "  ")
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

// Save writes the metrics to the specified file in JSON format.
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
	// Marshal the metrics to JSON format.
	logger.Log.Debugf("metrics saved (%d bytes) to %s", len(data), f.fname)
	return nil
}

// Load reads the metrics from the specified file and restores them to the storage.
func (f *FileSaver) Load(ctx context.Context) (map[string]float64, map[string]int64, error) {
	logger.Log.Debugf("loading metrics from %s", f.fname)
	// Check if the file name is empty -> not saving (used in tests).
	if f.fname == "" {
		return map[string]float64{}, map[string]int64{}, nil
	}

	// Check if the file exists.
	data, err := os.ReadFile(f.fname)
	if err != nil {
		return map[string]float64{}, map[string]int64{}, os.ErrNotExist
	}

	// Check if the data is empty.
	if len(data) == 0 {
		logger.Log.Warn("storage empty")
		return map[string]float64{}, map[string]int64{}, nil
	}

	// Unmarshal the data to a temporary struct.
	tmp := struct {
		Gauge   map[string]float64 `json:"gauge"`
		Counter map[string]int64   `json:"counter"`
	}{}

	if err := json.Unmarshal(data, &tmp); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			logger.Log.Warnf("data storage error: %v", err)
			return nil, nil, nil
		}
		return map[string]float64{}, map[string]int64{}, fmt.Errorf("error unmarshal metrics: %w", err)
	}

	logger.Log.Debugf("metrics restored from %s", f.fname)
	return tmp.Gauge, tmp.Counter, nil
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
