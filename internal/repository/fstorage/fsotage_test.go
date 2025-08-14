package fstorage

import (
	"context"
	"path/filepath"
	"testing"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	cfg "github.com/devize-ed/yapracproj-metrics.git/internal/repository/fstorage/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository/mstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tmpFilePath returns a temporary file path for testing
func tmpFilePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "metrics.json")
}

func TestFileSaver_SaveAndRestore(t *testing.T) {
	// Test data structures for gauge and counter metrics
	type Gauge struct {
		name  string
		value float64
	}
	type Counter struct {
		name  string
		delta int64
	}

	tests := []struct {
		name     string
		gauges   []Gauge
		counters []Counter
	}{
		{
			name:     "single_values",
			gauges:   []Gauge{{"pi", 3.14}},
			counters: []Counter{{"req_total", 42}},
		},
		{
			name: "multiple_values",
			gauges: []Gauge{
				{"temp", 25.5}, {"pressure", 1013.25},
			},
			counters: []Counter{
				{"requests", 100}, {"errors", 5},
			},
		},
		{
			name:     "empty_values",
			gauges:   []Gauge{},
			counters: []Counter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file path for this test
			filePath := tmpFilePath(t)

			// Create FileSaver with restore enabled
			config := &cfg.FStorageConfig{
				FPath:         filePath,
				StoreInterval: 0, // Synchronous save
				Restore:       true,
			}

			// Create the storage and file saver
			memStorage := mstorage.NewMemStorage()
			fs := NewFileSaver(context.Background(), config, memStorage)
			defer func() {
				require.NoError(t, fs.Close(), "failed to close file saver")
			}()

			// Set gauge values
			for _, g := range tt.gauges {
				err := fs.SetGauge(context.Background(), g.name, &g.value)
				require.NoError(t, err, "failed to set gauge %s", g.name)
			}

			// Add counter values
			for _, c := range tt.counters {
				err := fs.AddCounter(context.Background(), c.name, &c.delta)
				require.NoError(t, err, "failed to add counter %s", c.name)
			}

			// Force close to ensure data is saved
			require.NoError(t, fs.Close(), "failed to close file saver")

			// Create a new FileSaver to test restoration
			newMemStorage := mstorage.NewMemStorage()
			newFs := NewFileSaver(context.Background(), config, newMemStorage)
			defer func() {
				require.NoError(t, newFs.Close(), "failed to close file saver")
			}()

			// Verify gauge values were restored
			for _, g := range tt.gauges {
				val, err := newFs.GetGauge(context.Background(), g.name)
				require.NoError(t, err, "gauge %s should exist", g.name)
				assert.Equal(t, g.value, *val, "gauge %s value mismatch", g.name)
			}

			// Verify counter values were restored
			for _, c := range tt.counters {
				val, err := newFs.GetCounter(context.Background(), c.name)
				require.NoError(t, err, "counter %s should exist", c.name)
				assert.Equal(t, c.delta, *val, "counter %s value mismatch", c.name)
			}
		})
	}
}

func TestFileSaver_GetAllMetrics(t *testing.T) {
	filePath := tmpFilePath(t)
	config := &cfg.FStorageConfig{
		FPath:         filePath,
		StoreInterval: 0, // Synchronous save
		Restore:       false,
	}

	memStorage := mstorage.NewMemStorage()
	fs := NewFileSaver(context.Background(), config, memStorage)
	defer func() {
		require.NoError(t, fs.Close(), "failed to close file saver")
	}()

	// Add some test data
	gaugeVal := 42.5
	counterVal := int64(10)

	require.NoError(t, fs.SetGauge(context.Background(), "test_gauge", &gaugeVal))
	require.NoError(t, fs.AddCounter(context.Background(), "test_counter", &counterVal))

	// Get all metrics
	metrics, err := fs.GetAll(context.Background())
	require.NoError(t, err)

	// Verify the metrics are present
	assert.Contains(t, metrics, "test_gauge")
	assert.Contains(t, metrics, "test_counter")
	assert.Equal(t, "42.5", metrics["test_gauge"])
	assert.Equal(t, "10", metrics["test_counter"])
}

func TestFileSaver_SaveBatch(t *testing.T) {
	filePath := tmpFilePath(t)
	config := &cfg.FStorageConfig{
		FPath:         filePath,
		StoreInterval: 0, // Synchronous save
		Restore:       true,
	}

	memStorage := mstorage.NewMemStorage()
	fs := NewFileSaver(context.Background(), config, memStorage)
	defer func() {
		require.NoError(t, fs.Close(), "failed to close file saver")
	}()

	// Create a batch of metrics
	gaugeVal1 := 1.23
	gaugeVal2 := 4.56
	counterVal1 := int64(100)
	counterVal2 := int64(200)

	batch := []models.Metrics{
		{
			ID:    "gauge1",
			MType: models.Gauge,
			Value: &gaugeVal1,
		},
		{
			ID:    "gauge2",
			MType: models.Gauge,
			Value: &gaugeVal2,
		},
		{
			ID:    "counter1",
			MType: models.Counter,
			Delta: &counterVal1,
		},
		{
			ID:    "counter2",
			MType: models.Counter,
			Delta: &counterVal2,
		},
	}

	// Save the batch
	require.NoError(t, fs.SaveBatch(context.Background(), batch), "failed to save batch")

	// Force close to ensure data is persisted
	require.NoError(t, fs.Close(), "failed to close file saver")

	// Create a new FileSaver to test restoration
	newMemStorage := mstorage.NewMemStorage()
	newFs := NewFileSaver(context.Background(), config, newMemStorage)
	defer func() {
		require.NoError(t, newFs.Close(), "failed to close file saver")
	}()

	// Verify all metrics were saved and restored correctly
	val1, err := newFs.GetGauge(context.Background(), "gauge1")
	require.NoError(t, err)
	assert.Equal(t, gaugeVal1, *val1)

	val2, err := newFs.GetGauge(context.Background(), "gauge2")
	require.NoError(t, err)
	assert.Equal(t, gaugeVal2, *val2)

	val3, err := newFs.GetCounter(context.Background(), "counter1")
	require.NoError(t, err)
	assert.Equal(t, counterVal1, *val3)

	val4, err := newFs.GetCounter(context.Background(), "counter2")
	require.NoError(t, err)
	assert.Equal(t, counterVal2, *val4)
}
