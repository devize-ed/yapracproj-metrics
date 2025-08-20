package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func newTestAgent(host string) *Agent {
	client := resty.New()
	logger := zap.NewNop().Sugar()

	cfg := config.AgentConfig{}
	cfg.Connection.Host = host
	cfg.Agent.RateLimit = 4

	return NewAgent(client, cfg, logger)
}

func TestSendMetricsBatch_WithWorkerPool(t *testing.T) {
	var gotStatus int32
	var gotMethod string
	var gotPath string
	var reqCount int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		atomic.AddInt32(&reqCount, 1)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		atomic.StoreInt32(&gotStatus, http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	agent := newTestAgent(host)

	metrics := make([]models.Metrics, 0, batchSize*2+3)
	for i := 0; i < cap(metrics); i++ {
		val := float64(i)
		metrics = append(metrics, models.Metrics{
			ID:    "m" + strings.Repeat("x", i%3),
			MType: models.Gauge,
			Value: &val,
		})
	}

	numWorkers := 3
	j := NewJobs(numWorkers, agent.logger)
	errCh := j.createWorkerPool(agent.request, numWorkers, agent.logger)

	if err := j.sendMetricsBatch(agent.config.Connection.Host, metrics, agent.logger); err != nil {
		t.Fatalf("sendMetricsBatch() error = %v", err)
	}
	close(j.jobsQueue)
	j.wg.Wait()
	close(errCh)

	for err := range errCh {
		assert.NoError(t, err, "worker reported unexpected error")
	}

	assert.Equal(t, "/updates/", gotPath, "path should be /updates/")
	assert.Equal(t, "POST", gotMethod, "method should be POST")
	assert.Equal(t, http.StatusOK, int(atomic.LoadInt32(&gotStatus)))

	wantBatches := int32((len(metrics) + batchSize - 1) / batchSize)
	assert.Equal(t, wantBatches, atomic.LoadInt32(&reqCount), "unexpected number of POSTs (batches)")
}

func TestGetMetric_GaugeAndCounter(t *testing.T) {
	var gotStatus int
	var gotMethod string
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		gotStatus = http.StatusOK
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	agent := newTestAgent(host)

	if err := getMetric(agent.request, host, "testGauge", Gauge(0)); err != nil {
		t.Fatalf("getMetric(gauge) error = %v", err)
	}
	assert.Equal(t, "/value/", gotPath, "path should be /value/")
	assert.Equal(t, "POST", gotMethod, "method should be POST")
	assert.Equal(t, http.StatusOK, gotStatus)

	if err := getMetric(agent.request, host, "testCounter", Counter(1)); err != nil {
		t.Fatalf("getMetric(counter) error = %v", err)
	}
	assert.Equal(t, "/value/", gotPath, "path should be /value/")
	assert.Equal(t, "POST", gotMethod, "method should be POST")
	assert.Equal(t, http.StatusOK, gotStatus)
}

func TestCreateWorkerPool(t *testing.T) {
	var processed int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&processed, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	host := strings.TrimPrefix(srv.URL, "http://")

	agent := newTestAgent(host)

	numWorkers := 4
	j := NewJobs(numWorkers, agent.logger)
	errCh := j.createWorkerPool(agent.request, numWorkers, agent.logger)

	const N = 7
	for i := 0; i < N; i++ {
		j.jobsQueue <- batchRequest{
			name:      "job",
			endpoint:  "http://" + host + "/updates/",
			bodyBytes: []byte(`[]`),
		}
	}
	close(j.jobsQueue)
	j.wg.Wait()
	close(errCh)

	for err := range errCh {
		assert.NoError(t, err, "worker reported unexpected error")
	}
	assert.Equal(t, int32(N), atomic.LoadInt32(&processed), "not all jobs were processed")
}
