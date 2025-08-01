package fsaver

import (
	"context"
	"path/filepath"
	"testing"

	repository "github.com/devize-ed/yapracproj-metrics.git/internal/repository"
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
			pers := NewFileSaver(path)
			src := repository.NewMemStorage(0, pers)
			for _, g := range tt.gauge {
				src.SetGauge(context.Background(), g.k, g.v)
			}
			for _, c := range tt.counter {
				src.AddCounter(context.Background(), c.k, c.v)
			}

			require.NoError(t, src.SaveToRepo(context.Background()), "save failed")

			dst := repository.NewMemStorage(0, pers)
			require.NoError(t, dst.LoadFromRepo(context.Background()), "load failed")

			for _, g := range tt.gauge {
				val, ok := dst.GetGauge(context.Background(), g.k)
				assert.True(t, ok, "gauge %q missing", g.k)
				assert.Equal(t, g.v, val)
			}
			for _, c := range tt.counter {
				val, ok := dst.GetCounter(context.Background(), c.k)
				assert.True(t, ok, "counter %q missing", c.k)
				assert.Equal(t, c.v, val)
			}
		})
	}
}
