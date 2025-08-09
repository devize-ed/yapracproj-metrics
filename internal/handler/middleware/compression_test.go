package handler

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareGzip(t *testing.T) {
	_ = logger.Initialize("debug")
	defer logger.Log.Sync()

	requestBody := `{
		"id":"LastGC",
		"type":"gauge"
	}`
	successBody := `{
		"id":"LastGC",
		"type":"gauge",
		"value":1744184459
	}`

	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(successBody))
	})

	router := chi.NewRouter()
	router.Use(MiddlewareGzip)
	router.Post("/", successHandler)

	srv := httptest.NewServer(router)
	defer srv.Close()
	client := resty.New()

	t.Run("client_sends_gzip", func(t *testing.T) {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		_, err := zw.Write([]byte(requestBody))
		require.NoError(t, err)
		require.NoError(t, zw.Close())

		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Accept-Encoding", "").
			SetBody(buf.Bytes()).
			Post(srv.URL + "/")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		require.Empty(t, resp.Header().Get("Content-Encoding"))
		require.JSONEq(t, successBody, string(resp.Body()))
	})

	t.Run("server_sends_gzip", func(t *testing.T) {
		resp, err := client.R().
			SetHeader("Accept-Encoding", "gzip").
			SetBody([]byte(requestBody)).
			Post(srv.URL + "/")
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.JSONEq(t, successBody, string(resp.Body()))
	})
}
