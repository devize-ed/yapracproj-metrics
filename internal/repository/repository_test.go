// repository/storage_test.go
package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStorage() *MemStorage {
	return NewMemStorage(0, NewStubRepository())
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
				ms.SetGauge("testMetric", 123.456)
				return ms
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ms.SetGauge(tt.metricName, tt.value)

			got, ok := tt.ms.GetGauge(tt.metricName)
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
	ms.SetGauge("testMetric", 123.456)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ms.GetGauge(tt.metricName)

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
			tt.ms.AddCounter(tt.metricName, tt.delta)

			got, ok := tt.ms.GetCounter(tt.metricName)
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
			got, ok := ms.GetCounter(tt.metricName)

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
		st := NewMemStorage(1, NewFilePersister(path)) // without save

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		st.IntervalSaver(ctx, 1)

		st.SetGauge("TestGauge", 123.11)

		time.Sleep(2 * time.Second) // wait for the ticker to tick

		_, err := os.Stat(path)
		assert.NoError(t, err, "file should be created after the interval")

		cancel()
		time.Sleep(50 * time.Millisecond)

		check := NewMemStorage(0, NewFilePersister(path))
		require.NoError(t, check.Load())
		val, ok := check.GetGauge("TestGauge")
		assert.True(t, ok)
		assert.Equal(t, 123.11, val)
	})

	t.Run("sync_save", func(t *testing.T) {
		path := tmpFilePath(t)
		st := NewMemStorage(0, NewFilePersister(path)) //turn on sync save

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		st.IntervalSaver(ctx, 0) // sync mode

		st.SetGauge("TestGauge", 123.4)

		_, err := os.Stat(path)
		require.NoError(t, err, "file should be created for sync save")

		load := NewMemStorage(0, NewFilePersister(path))
		require.NoError(t, load.Load())
		v, ok := load.GetGauge("TestGauge")
		assert.True(t, ok)
		assert.Equal(t, 123.4, v)

		cancel()
		time.Sleep(1 * time.Second) // wait for the goroutine to finish
	})
}
