// repository/storage_test.go
package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
