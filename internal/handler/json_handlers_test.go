package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	storage "github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestUpdateJsonHandler(t *testing.T) {
	_ = logger.Initialize("debug")
	defer logger.Log.Sync()

	endpoint := "/update"
	ms := storage.NewMemStorage(0, storage.NewStubStorage())
	h := NewHandler(ms)

	r := chi.NewRouter()
	r.Post(endpoint, h.UpdateMetricJSONHandler())

	srv := httptest.NewServer(r)
	defer srv.Close()

	var tests = []struct {
		name                string
		expectedContentType string
		expectedCode        int
		body                string
	}{
		{
			name:                "update_counter",
			expectedContentType: "application/json",
			expectedCode:        http.StatusOK,
			body:                `{"id": "testCounter","type": "counter","delta": 5}`,
		},
		{
			name:                "update_gauge",
			expectedContentType: "application/json",
			expectedCode:        http.StatusOK,
			body:                `{"id": "testGauge1","type": "gauge","value": 123123}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL = srv.URL + endpoint

			req.SetHeader("Content-Type", "application/json")
			req.SetBody(tt.body)

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			assert.Equal(t, tt.expectedContentType, resp.Header().Get("Content-Type"), "Response content type didn't match expected")
		})
	}
}

func TestHandler_GetMetricJsonHandler(t *testing.T) {
	_ = logger.Initialize("debug")
	defer logger.Log.Sync()

	endpoint := "/value"
	ms := storage.NewMemStorage(0, storage.NewStubStorage())
	h := NewHandler(ms)
	testMemoryStorage(ms)

	r := chi.NewRouter()
	r.Post(endpoint, h.GetMetricJSONHandler())

	srv := httptest.NewServer(r)
	defer srv.Close()

	var tests = []struct {
		name                string
		expectedContentType string
		expectedCode        int
		body                string
		expectedBody        string
	}{
		{
			name:                "get_counter",
			expectedContentType: "application/json",
			expectedCode:        http.StatusOK,
			body:                `{"id": "testCounter","type": "counter"}`,
			expectedBody:        `{"id": "testCounter","type": "counter","delta": 5}`,
		},
		{
			name:                "get_gauge",
			expectedContentType: "application/json",
			expectedCode:        http.StatusOK,
			body:                `{"id": "testGauge1","type": "gauge"}`,
			expectedBody:        `{"id": "testGauge1","type": "gauge","value": 10.5}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL = srv.URL + endpoint

			req.SetHeader("Content-Type", "application/json")
			req.SetBody(tt.body)

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			assert.Equal(t, tt.expectedContentType, resp.Header().Get("Content-Type"), "Response content type didn't match expected")
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, string(resp.Body()), "Response body didn't match expected")
			}
		})
	}
}
