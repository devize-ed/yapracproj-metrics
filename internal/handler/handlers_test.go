package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	mstorage "github.com/devize-ed/yapracproj-metrics.git/internal/repository/mstorage"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func testMemoryStorage(t *testing.T, ms *mstorage.MemStorage) {
	ctx := context.Background()
	delta := int64(5)
	if err := ms.AddCounter(ctx, "testCounter", &delta); err != nil {
		t.Fatalf("Failed to add counter: %v", err)
	}
	g1 := 10.5
	if err := ms.SetGauge(ctx, "testGauge1", &g1); err != nil {
		t.Fatalf("Failed to set gauge: %v", err)
	}
	g2 := 1.5
	if err := ms.SetGauge(ctx, "testGauge2", &g2); err != nil {
		t.Fatalf("Failed to set gauge: %v", err)
	}
}

func testRequest(t *testing.T, srv *httptest.Server, method, path string) (resp *resty.Response) {
	req := resty.New().R()
	req.Method = method
	req.URL = srv.URL + path

	resp, err := req.Send()
	assert.NoError(t, err, "error making HTTP request")
	return resp
}

func TestUpdateHandler(t *testing.T) {
	logger, err := logger.Initialize("debug")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	ms := mstorage.NewMemStorage()
	h := NewHandler(ms, "", logger)

	r := chi.NewRouter()
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateMetricHandler())

	srv := httptest.NewServer(r)
	defer srv.Close()

	var tests = []struct {
		url                 string
		expectedContentType string
		expectedCode        int
	}{
		{"/update/counter/testCounter/123", "text/html; charset=utf-8", http.StatusOK},
		{"/update/gauge/testGauge/123", "text/html; charset=utf-8", http.StatusOK},
		{"/update/gauge/", "text/plain; charset=utf-8", http.StatusNotFound},
		{"/update/incorrectMetricType/testMetric/123", "text/plain; charset=utf-8", http.StatusBadRequest},
		{"/update/counter/testCounter/stringValue", "text/plain; charset=utf-8", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL = srv.URL + tt.url
			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			assert.Equal(t, tt.expectedContentType, resp.Header().Get("Content-Type"), "Response content type didn't match expected")
		})
	}
}

func TestListAllHandler(t *testing.T) {
	logger, err := logger.Initialize("debug")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	ms := mstorage.NewMemStorage()
	h := NewHandler(ms, "", logger)
	testMemoryStorage(t, ms)

	r := chi.NewRouter()
	r.Get("/", h.ListMetricsHandler())
	srv := httptest.NewServer(r)
	defer srv.Close()

	var tests = []struct {
		url          string
		expectedCode int
		expectedBody string
	}{
		{"/", http.StatusOK, "testCounter = 5\ntestGauge1 = 10.5\ntestGauge2 = 1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			resp := testRequest(t, srv, http.MethodGet, tt.url)
			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			assert.Equal(t, tt.expectedBody, resp.String(), "Response body didn't match expected")
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	logger, err := logger.Initialize("debug")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	ms := mstorage.NewMemStorage()
	h := NewHandler(ms, "", logger)
	testMemoryStorage(t, ms)

	r := chi.NewRouter()
	r.Get("/value/{metricType}/{metricName}", h.GetMetricHandler())

	srv := httptest.NewServer(r)
	defer srv.Close()

	var tests = []struct {
		url          string
		expectedCode int
		expectedBody string
	}{
		{"/value/counter/testCounter", http.StatusOK, "5"},
		{"/value/gauge/testGauge1", http.StatusOK, "10.5"},
		{"/value/gauge/", http.StatusNotFound, "404 page not found"},
		{"/value/gauge/unknownMetric", http.StatusNotFound, "metric not found"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			resp := testRequest(t, srv, http.MethodGet, tt.url)
			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			assert.Equal(t, tt.expectedBody, resp.String(), "Response body didn't match expected")
		})

	}
}
