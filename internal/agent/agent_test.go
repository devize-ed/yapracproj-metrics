package agent

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"

// 	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
// 	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
// 	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
// 	"github.com/go-resty/resty/v2"
// 	"github.com/stretchr/testify/assert"
// )

// func TestSendMetricsBatch(t *testing.T) {
// 	_ = logger.Initialize("debug")
// 	defer logger.SafeSync()

// 	type args struct {
// 		metric string
// 		value  Gauge
// 	}
// 	tests := []struct {
// 		name     string
// 		args     args
// 		wantErr  bool
// 		wantCode int
// 	}{
// 		{
// 			name: "send_metric",
// 			args: args{
// 				metric: "testMetric",
// 				value:  111.11,
// 			},
// 			wantErr:  false,
// 			wantCode: http.StatusOK,
// 		},
// 	}

// 	var gotStatus int
// 	var gotPath string
// 	var gotMethod string

// 	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		gotMethod = r.Method
// 		gotPath = r.URL.Path
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusOK)
// 		gotStatus = http.StatusOK
// 		_, _ = w.Write([]byte(`{"status":"ok"}`))
// 	}))
// 	defer srv.Close()

// 	host := strings.TrimPrefix(srv.URL, "http://")
// 	client := resty.New()
// 	cfg := config.AgentConfig{}
// 	cfg.Connection.Host = host

// 	agent := NewAgent(client, cfg)
// 	// Ensure SendMetricsBatch can enqueue without blocking.
// 	agent.jobsQueue = make(chan batchRequest, 1)

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := agent.SendMetricsBatch([]models.Metrics{
// 				{
// 					ID:    tt.args.metric,
// 					MType: "gauge",
// 					Value: func() *float64 { v := float64(tt.args.value); return &v }(),
// 				},
// 			})
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("SendMetricsBatch() error = %v, wantErr %w", err, tt.wantErr)
// 			}

// 			// Manually process the queued job (minimal change, no worker pool).
// 			job := <-agent.jobsQueue
// 			callErr := agent.Request(job.name, job.endpoint, job.bodyBytes)
// 			if (callErr != nil) != tt.wantErr {
// 				t.Errorf("Request() error = %v, wantErr %v", callErr, tt.wantErr)
// 			}

// 			assert.Equal(t, "/updates/", gotPath, "path should be /updates/")
// 			assert.Equal(t, "POST", gotMethod, "method should be POST")
// 			assert.Equal(t, tt.wantCode, gotStatus, "Expected status code to be %d", tt.wantCode)
// 		})
// 	}
// }

// func TestGetMetric(t *testing.T) {
// 	_ = logger.Initialize("debug")
// 	defer logger.SafeSync()

// 	type args struct {
// 		metric string
// 	}
// 	tests := []struct {
// 		name     string
// 		args     args
// 		wantErr  bool
// 		wantCode int
// 	}{
// 		{
// 			name: "get_metric",
// 			args: args{
// 				metric: "testMetric",
// 			},
// 			wantErr:  false,
// 			wantCode: http.StatusOK,
// 		},
// 	}

// 	var gotStatus int
// 	var gotPath string
// 	var gotMethod string

// 	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		gotMethod = r.Method
// 		gotPath = r.URL.Path
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusOK)
// 		gotStatus = http.StatusOK
// 		_, _ = w.Write([]byte(`{"status":"ok"}`))
// 	}))
// 	defer srv.Close()

// 	host := strings.TrimPrefix(srv.URL, "http://")
// 	client := resty.New()
// 	cfg := config.AgentConfig{}
// 	cfg.Connection.Host = host

// 	agent := NewAgent(client, cfg)
// 	// Enable gzip path coverage without changing test shape (optional, harmless).
// 	agent.config.Agent.EnableGzip = true

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := GetMetric(agent, tt.args.metric, Gauge(0)); (err != nil) != tt.wantErr {
// 				t.Errorf("GetMetric() error = %v, wantErr %w", err, tt.wantErr)
// 			}
// 			assert.Equal(t, "/value/", gotPath, "path should be /value/")
// 			assert.Equal(t, "POST", gotMethod, "method should be POST")
// 			assert.Equal(t, tt.wantCode, gotStatus, "Expected status code to be %d", tt.wantCode)
// 		})
// 	}
// }
