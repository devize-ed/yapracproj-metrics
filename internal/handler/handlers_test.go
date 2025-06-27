package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	storage "github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func testMemoryStorage(ms *storage.MemStorage) {
	ms.AddCounter("testCounter", 5)
	ms.SetGauge("testGauge1", 10.5)
	ms.SetGauge("testGauge2", 10.5)
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
	ms := storage.NewMemStorage()
	r := chi.NewRouter()
	r.Post("/update/{metricType}/{metricName}/{metricValue}", UpdateMetricHandler(ms))

	srv := httptest.NewServer(r)
	defer srv.Close()

	var tests = []struct {
		url                 string
		expectedContentType string
		expectedCode        int
	}{
		{"/update/counter/testCounter/123", "text/plain; charset=utf-8", http.StatusOK},
		{"/update/gauge/testGauge/123", "text/plain; charset=utf-8", http.StatusOK},
		{"/update/gauge/", "text/plain; charset=utf-8", http.StatusNotFound},
		{"/update/incorrectMetricType/testMetric/123", "text/plain; charset=utf-8", http.StatusBadRequest},
		{"/update/counter/testCounter/stringValue", "text/plain; charset=utf-8", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL = srv.URL + tt.url
			fmt.Println("Request URL:", req.URL)
			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			assert.Equal(t, tt.expectedContentType, resp.Header().Get("Content-Type"), "Response content type didn't match expected")
		})
	}
}

func TestListAllHandler(t *testing.T) {
	ms := storage.NewMemStorage()
	testMemoryStorage(ms)

	r := chi.NewRouter()
	r.Get("/", ListAllHandler(ms))
	srv := httptest.NewServer(r)
	defer srv.Close()

	var tests = []struct {
		url          string
		expectedCode int
		expectedBody string
	}{
		{"/", http.StatusOK, "testCounter = 5\ntestGauge1 = 10.5\ntestGauge2 = 10.5"},
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
	ms := storage.NewMemStorage()
	testMemoryStorage(ms)

	r := chi.NewRouter()
	r.Get("/value/{metricType}/{metricName}", GetMetricHandler(ms))

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
