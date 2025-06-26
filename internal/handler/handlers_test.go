package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/repository/storage"
	"github.com/stretchr/testify/assert"
)

func TestUpdateHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "successful update of counter",
			request: "/update/counter/testMetric/123",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  200,
			},
		},
		{
			name:    "successful update of gauge",
			request: "/update/gauge/testMetric/123",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  200,
			},
		},
		{
			name:    "empty metric name",
			request: "/update/gauge/",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
			},
		},
		{
			name:    "incorrect mettric type",
			request: "/update/incorrectMettricType/testMetric/123",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
			},
		},
		{
			name:    "empty metric name",
			request: "/update/counter/testMetric/stringValue",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
			},
		},
	}

	ms := storage.NewMemStorage()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(UpdateHandler(ms))
			h(w, request)
			res := w.Result()
			defer res.Body.Close()
			io.Copy(io.Discard, res.Body)

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}

func TestMiddleware(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name   string
		method string
		want   want
	}{
		{
			name:   "POST method",
			method: http.MethodPost,
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "GET method",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "DELETE method",
			method: http.MethodDelete,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, "/update/counter/testMetric/123", nil)
			w := httptest.NewRecorder()

			Middleware(next).ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()
			io.Copy(io.Discard, res.Body)

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
		})
	}
}
