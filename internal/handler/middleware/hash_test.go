package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/devize-ed/yapracproj-metrics.git/internal/sign"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestHashMiddleware(t *testing.T) {
	_ = logger.Initialize("debug")
	defer logger.Log.Sync()

	requestBody := `{
		"id":"LastGC",
		"type":"gauge"
	}`

	key := "test_key"

	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	router := chi.NewRouter()
	router.Use(HashMiddleware(key))
	router.Post("/", successHandler)

	srv := httptest.NewServer(router)
	defer srv.Close()
	client := resty.New()

	tests := []struct {
		name       string
		hash       string
		key        string
		wantStatus int
		wantBody   string
	}{
		{name: "request_with_hash",
			hash:       sign.Hash([]byte(requestBody), key),
			key:        key,
			wantStatus: http.StatusOK,
			wantBody:   "success",
		},
		{name: "missing_hash",
			hash:       "",
			key:        key,
			wantStatus: http.StatusBadRequest,
			wantBody:   "missing hash header\n",
		},
		{name: "invalid_hash",
			hash:       "1d23d23d231",
			key:        key,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Hash verification failed\n",
		},
		{name: "empty_key",
			hash:       sign.Hash([]byte(requestBody), ""),
			key:        "",
			wantStatus: http.StatusBadRequest,
			wantBody:   "Hash verification failed\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader(sign.HashHeader, test.hash).
				SetBody([]byte(requestBody)).
				Post(srv.URL + "/")
			require.NoError(t, err)
			require.Equal(t, test.wantStatus, resp.StatusCode())
			require.Equal(t, test.wantBody, string(resp.Body()))
		})
	}
}
