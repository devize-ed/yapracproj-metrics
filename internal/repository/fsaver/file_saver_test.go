package fsaver

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tmpFilePath returns a temporary file path for testing
func tmpFilePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "metrics.json")
}

func TestStorage_SaveAndLoad(t *testing.T) {
	// type Gauge is used for gauge metrics
	type Gauge struct {
		k string
		v float64
	}
	// type Counter is used for counter metrics
	type Counter struct {
		k string
		v int64
	}

	tests := []struct {
		name    string
		gauge   []Gauge
		counter []Counter
	}{
		{
			name:    "single_values",
			gauge:   []Gauge{{"pi", 3.14}},
			counter: []Counter{{"req_total", 42}},
		},
		{
			name: "multiple_values",
			gauge: []Gauge{
				{"a", 1.1}, {"b", 2.2},
			},
			counter: []Counter{
				{"x", 10}, {"y", 20},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tmpFilePath(t)

			src := NewMemStorage(0, )
			for _, g := range tt.gauge {
				src.SetGauge(g.k, g.v)
			}
			for _, c := range tt.counter {
				src.AddCounter(c.k, c.v)
			}

			require.NoError(t, src.Save(path), "save failed")

			dst := NewMemStorage(0, path)
			require.NoError(t, dst.Load(path), "load failed")

			for _, g := range tt.gauge {
				val, ok := dst.GetGauge(g.k)
				assert.True(t, ok, "gauge %q missing", g.k)
				assert.Equal(t, g.v, val)
			}
			for _, c := range tt.counter {
				val, ok := dst.GetCounter(c.k)
				assert.True(t, ok, "counter %q missing", c.k)
				assert.Equal(t, c.v, val)
			}
		})
	}
}

func TestStorage_IntervalSaver(t *testing.T) {
	t.Run("periodic_save", func(t *testing.T) {
		path := tmpFilePath(t)
		st := NewMemStorage(1, path) // without save

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		st.IntervalSaver(ctx, 1, path)

		st.SetGauge("TestGauge", 123.11)

		time.Sleep(2 * time.Second) // wait for the ticker to tick

		_, err := os.Stat(path)
		assert.NoError(t, err, "file should be created after the interval")

		cancel()
		time.Sleep(50 * time.Millisecond)

		check := NewMemStorage(0, path)
		require.NoError(t, check.Load(path))
		val, ok := check.GetGauge("TestGauge")
		assert.True(t, ok)
		assert.Equal(t, 123.11, val)
	})

	t.Run("sync_save", func(t *testing.T) {
		path := tmpFilePath(t)
		st := NewMemStorage(0, path) //turn on sync save

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		st.IntervalSaver(ctx, 0, path) // sync mode

		st.SetGauge("TestGauge", 123.4)

		_, err := os.Stat(path)
		require.NoError(t, err, "file should be created for sync save")

		load := NewMemStorage(0, path)
		require.NoError(t, load.Load(path))
		v, ok := load.GetGauge("TestGauge")
		assert.True(t, ok)
		assert.Equal(t, 123.4, v)

		cancel()
		time.Sleep(1 * time.Second) // wait for the goroutine to finish
	})
}
