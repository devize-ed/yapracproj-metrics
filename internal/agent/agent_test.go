package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestSendMetric(t *testing.T) {
	_ = logger.Initialize("debug")
	defer logger.Log.Sync()

	type args struct {
		metric string
		value  Gauge
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantCode int
	}{
		{
			name: "send_metric",
			args: args{
				metric: "testMetric",
				value:  111.11,
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
	cfg := config.AgentConfig{Host: host}

	agent := NewAgent(client, cfg)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendMetric(agent, tt.args.metric, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("SendMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		assert.Equal(t, tt.wantCode, http.StatusOK, "Expected status code to be %d", tt.wantCode)
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
	cfg := config.AgentConfig{Host: host}

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
