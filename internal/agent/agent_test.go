package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestSendMetricsBatch(t *testing.T) {
	_ = logger.Initialize("debug")
	defer logger.Log.Sync()

	tests := []struct {
		name     string
		metrics  []models.Metrics
		wantErr  bool
		wantCode int
	}{
		{
			name: "send_metrics_batch",
			metrics: []models.Metrics{
				{
					ID:    "testMetric",
					MType: "gauge",
					Value: func() *float64 { v := 111.11; return &v }(),
				},
			},
			wantErr:  false,
			wantCode: http.StatusOK,
		},
	}

	var gotStatusCode int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		gotStatusCode = http.StatusOK
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	client := resty.New()
	cfg := config.AgentConfig{}
	cfg.Connection.Host = host

	agent := NewAgent(client, cfg)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatusCode = 0
			err := SendMetricsBatch(agent, tt.metrics)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendMetricsBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantCode, gotStatusCode, "Expected status code to be %d", tt.wantCode)
		})
	}
}

func TestGetMetric(t *testing.T) {
	_ = logger.Initialize("debug")
	defer logger.Log.Sync()

	type args struct {
		metric string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantCode int
	}{
		{
			name: "get_metric",
			args: args{
				metric: "testMetric",
			},
			wantErr:  false,
			wantCode: http.StatusOK,
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	client := resty.New()
	cfg := config.AgentConfig{}
	cfg.Connection.Host = host

	agent := NewAgent(client, cfg)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GetMetric(agent, tt.args.metric, Gauge(0)); (err != nil) != tt.wantErr {
				t.Errorf("GetMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		assert.Equal(t, tt.wantCode, http.StatusOK, "Expected status code to be %d", tt.wantCode)
	}
}
