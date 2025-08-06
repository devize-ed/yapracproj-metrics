package mstorage

import (
	"context"
	"testing"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStorage() *MemStorage {
	return NewMemStorage()
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
				val := 123.456
				require.NoError(t, ms.SetGauge(context.Background(), "testMetric", &val))
				return ms
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := tt.value
			require.NoError(t, tt.ms.SetGauge(context.Background(), tt.metricName, &val))

			got, err := tt.ms.GetGauge(context.Background(), tt.metricName)
			require.NoError(t, err)
			assert.Equal(t, tt.want, *got)
		})
	}
}

// TestMemStorage_GetGauge verifies retrieval of gauge metrics.
func TestMemStorage_GetGauge(t *testing.T) {
	ms := newTestStorage()
	v := 123.456
	require.NoError(t, ms.SetGauge(context.Background(), "testMetric", &v))

	tests := []struct {
		name       string
		metricName string
		wantValue  float64
		wantErr    bool
	}{
		{
			name:       "successful get gauge",
			metricName: "testMetric",
			wantValue:  123.456,
			wantErr:    false,
		},
		{
			name:       "gauge not found",
			metricName: "unknownMetric",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ms.GetGauge(context.Background(), tt.metricName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantValue, *got)
			}
		})
	}
}

// TestMemStorage_AddCounter verifies that counters can be added and incremented.
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
			d := tt.delta
			require.NoError(t, tt.ms.AddCounter(context.Background(), tt.metricName, &d))

			got, err := tt.ms.GetCounter(context.Background(), tt.metricName)
			require.NoError(t, err)
			assert.Equal(t, tt.want, *got)
		})
	}
}

// TestMemStorage_GetCounter verifies retrieval of counter metrics.
func TestMemStorage_GetCounter(t *testing.T) {
	ms := newTestStorage()
	ms.Counter["testMetric"] = 5

	tests := []struct {
		name       string
		metricName string
		wantValue  int64
		wantErr    bool
	}{
		{
			name:       "successful get counter",
			metricName: "testMetric",
			wantValue:  5,
			wantErr:    false,
		},
		{
			name:       "counter not found",
			metricName: "unknownMetric",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ms.GetCounter(context.Background(), tt.metricName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantValue, *got)
			}
		})
	}
}

// TestMemStorage_SaveBatchAndGetAll verifies batch saving and retrieval of all metrics.
func TestMemStorage_SaveBatchAndGetAll(t *testing.T) {
	ms := newTestStorage()

	tGauge1 := 123.11
	tGauge2 := 11.123
	tCounter1 := int64(5)
	batch := []models.Metrics{
		{ID: "testGauge1", MType: models.Gauge, Value: &tGauge1},
		{ID: "testGauge2", MType: models.Gauge, Value: &tGauge2},
		{ID: "testCounter1", MType: models.Counter, Delta: &tCounter1},
	}
	expected := map[string]string{
		"testGauge1":   "123.11",
		"testGauge2":   "11.123",
		"testCounter1": "5",
	}

	require.NoError(t, ms.SaveBatch(context.Background(), batch))

	g, err := ms.GetGauge(context.Background(), "testGauge1")
	require.NoError(t, err)
	assert.Equal(t, tGauge1, *g)

	c, err := ms.GetCounter(context.Background(), "testCounter1")
	require.NoError(t, err)
	assert.Equal(t, tCounter1, *c)

	all, err := ms.GetAll(context.Background())
	require.NoError(t, err)
	assert.Equal(t, expected, all)
}
