package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/audit"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	mstorage "github.com/devize-ed/yapracproj-metrics.git/internal/repository/mstorage"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestUpdateJsonHandler(t *testing.T) {
	logger, err := logger.Initialize("debug")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	endpoint := "/update"
	ms := mstorage.NewMemStorage()
	auditor := audit.NewAuditor(logger, "", "")
	h := NewHandler(ms, "", auditor, "", logger)

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
	logger, err := logger.Initialize("debug")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	endpoint := "/value"
	ms := mstorage.NewMemStorage()
	auditor := audit.NewAuditor(logger, "", "")
	h := NewHandler(ms, "", auditor, "", logger)
	testMemoryStorage(t, ms)

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

// Examples:
func Example_jsonParams() {
	// Initialize logger
	log, err := logger.Initialize("debug")
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer func() {
		_ = log.Sync()
	}()

	// Initialize endpoints
	endpointUpdate := "/update"
	endpointGet := "/value"

	// Initialize storage and auditor
	ms := mstorage.NewMemStorage()
	auditor := audit.NewAuditor(log, "", "")

	// Initialize handler
	h := NewHandler(ms, "", auditor, "", log)

	// Initialize router
	r := chi.NewRouter()
	r.Post(endpointUpdate, h.UpdateMetricJSONHandler())
	r.Post(endpointGet, h.GetMetricJSONHandler())

	// Initialize server
	srv := httptest.NewServer(r)
	defer srv.Close()

	// Initialize client
	client := resty.New()

	// Update counter metric
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"id":"testCounter","type":"counter","delta":5}`).
		Post(srv.URL + endpointUpdate)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("Response code (update counter):", resp.StatusCode())

	// Update gauge metric
	resp, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"id":"testGauge1","type":"gauge","value":123123}`).
		Post(srv.URL + endpointUpdate)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("Response code (update gauge):", resp.StatusCode())

	// Get counter metric
	resp, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"id":"testCounter","type":"counter"}`).
		Post(srv.URL + endpointGet)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("Response code (get counter):", resp.StatusCode())
	fmt.Println("Response body (get counter):", string(resp.Body()))

	// Output:
	// Response code (update counter): 200
	// Response code (update gauge): 200
	// Response code (get counter): 200
	// Response body (get counter): {"id":"testCounter","type":"counter","delta":5}
}

func ExampleHandler_UpdateBatchHandler() {
	// Initialize logger
	log, err := logger.Initialize("debug")
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer func() {
		_ = log.Sync()
	}()

	// Initialize endpoints
	endpoint := "/updates"

	// Initialize storage and auditor
	ms := mstorage.NewMemStorage()
	auditor := audit.NewAuditor(log, "", "")

	// Initialize handler
	h := NewHandler(ms, "", auditor, "", log)

	// Initialize router
	r := chi.NewRouter()
	r.Post(endpoint, h.UpdateBatchHandler())

	// Initialize server
	srv := httptest.NewServer(r)
	defer srv.Close()

	// Initialize client
	client := resty.New()

	// Update batch metric
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`[{"id":"testCounter","type":"counter","delta":5},{"id":"testGauge1","type":"gauge","value":123123}]`).
		Post(srv.URL + endpoint)
	if err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println("Response code:", resp.StatusCode())
	fmt.Println("Response body:", string(resp.Body()))

	// Output:
	// Response code: 200
	// Response body: {"status":"ok"}
}
