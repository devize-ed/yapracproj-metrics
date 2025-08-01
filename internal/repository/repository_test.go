package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	fsaver "github.com/devize-ed/yapracproj-metrics.git/internal/repository/fsaver"
)

func tmpFilePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "metrics.json")
}

func newTestStorage() *MemStorage {
	return NewMemStorage(0, NewStubStorage())
}

func TestMemStorage_SetGauge(t *testing.T) {
	tests := []struct {
		name       string
		ms         *MemStorage
		metricName string
		value      float64
		want       float64
	}{
		{
			name:       "successful set gauge",
			ms:         newTestStorage(),
			metricName: "testMetric",
			value:      123.456,
			want:       123.456,
		},
		{
			name:       "override existing gauge",
			metricName: "testMetric",
			value:      123,
			want:       123,
			ms: func() *MemStorage {
				ms := newTestStorage()
				ms.SetGauge(context.Background(), "testMetric", 123.456)
				return ms
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ms.SetGauge(context.Background(), tt.metricName, tt.value)

			got, ok := tt.ms.GetGauge(context.Background(), tt.metricName)
			assert.True(t, ok, "metric should exist")
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_GetGauge(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		wantValue  float64
		wantOK     bool
	}{
		{
			name:       "successful get gauge",
			metricName: "testMetric",
			wantValue:  123.456,
			wantOK:     true,
		},
		{
			name:       "gauge not found",
			metricName: "unknownMetric",
			wantValue:  0,
			wantOK:     false,
		},
	}

	ms := newTestStorage()
	ms.SetGauge(context.Background(), "testMetric", 123.456)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ms.GetGauge(context.Background(), tt.metricName)

			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.wantValue, got)
			}
		})
	}
}

func TestMemStorage_AddCounter(t *testing.T) {
	tests := []struct {
		name       string
		ms         *MemStorage
		metricName string
		delta      int64
		want       int64
	}{
		{
			name:       "new counter metric",
			ms:         newTestStorage(),
			metricName: "testMetric",
			delta:      5,
			want:       5,
		},
		{
			name:       "increment existing counter",
			metricName: "testMetric",
			delta:      5,
			want:       15,
			ms: func() *MemStorage {
				ms := newTestStorage()
				ms.Counter["testMetric"] = 10
				return ms
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ms.AddCounter(context.Background(), tt.metricName, tt.delta)

			got, ok := tt.ms.GetCounter(context.Background(), tt.metricName)
			assert.True(t, ok, "metric should exist")
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_GetCounter(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		wantValue  int64
		wantOK     bool
	}{
		{
			name:       "successful get counter",
			metricName: "testMetric",
			wantValue:  5,
			wantOK:     true,
		},
		{
			name:       "counter not found",
			metricName: "unknownMetric",
			wantValue:  0,
			wantOK:     false,
		},
	}

	ms := newTestStorage()
	ms.Counter["testMetric"] = 5

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ms.GetCounter(context.Background(), tt.metricName)

			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.wantValue, got)
			}
		})
	}
}

func TestStorage_IntervalSaver(t *testing.T) {
	t.Run("periodic_save", func(t *testing.T) {
		path := tmpFilePath(t)
		st := NewMemStorage(1, fsaver.NewFileSaver(path)) // without save

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		st.IntervalSaver(ctx, 1)

		st.SetGauge(context.Background(), "TestGauge", 123.11)

		time.Sleep(2 * time.Second) // wait for the ticker to tick

		_, err := os.Stat(path)
		assert.NoError(t, err, "file should be created after the interval")

		cancel()
		time.Sleep(50 * time.Millisecond)

		check := NewMemStorage(0, fsaver.NewFileSaver(path))
		require.NoError(t, check.LoadFromRepo(context.Background()))
		val, ok := check.GetGauge(context.Background(), "TestGauge")
		assert.True(t, ok)
		assert.Equal(t, 123.11, val)
	})

	t.Run("sync_save", func(t *testing.T) {
		path := tmpFilePath(t)
		st := NewMemStorage(0, fsaver.NewFileSaver(path)) //turn on sync save

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		st.IntervalSaver(ctx, 0) // sync mode

		st.SetGauge(context.Background(), "TestGauge", 123.4)

		_, err := os.Stat(path)
		require.NoError(t, err, "file should be created for sync save")

		load := NewMemStorage(0, fsaver.NewFileSaver(path))
		require.NoError(t, load.LoadFromRepo(context.Background()))
		v, ok := load.GetGauge(context.Background(), "TestGauge")
		assert.True(t, ok)
		assert.Equal(t, 123.4, v)

		cancel()
		time.Sleep(1 * time.Second) // wait for the goroutine to finish
	})
}
