package handler

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPFilterMiddleware(t *testing.T) {
	logger, err := logger.Initialize("debug")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	trustedSubnet := "192.168.1.0/24"

	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})
	router := chi.NewRouter()
	router.Use(IPFilterMiddleware(IPFilterConfig{TrustedSubnet: trustedSubnet}, logger))

	tests := []struct {
		name       string
		ip         string
		key        string
		wantStatus int
		wantBody   string
	}{
		{name: "request_with_ip",
			ip:         "192.168.1.1",
			wantStatus: http.StatusOK,
			wantBody:   "success",
		},
		{name: "missing_ip",
			ip:         "",
			wantStatus: http.StatusForbidden,
			wantBody:   "IP address empty\n",
		},
		{name: "invalid_ip",
			ip:         "168.1.1.1",
			wantStatus: http.StatusForbidden,
			wantBody:   "IP address not in trusted subnet\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router.Post("/", successHandler)

			srv := httptest.NewServer(router)
			defer srv.Close()
			client := resty.New()

			req := client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("X-Real-IP", test.ip)
			resp, err := req.Post(srv.URL + "/")
			require.NoError(t, err)
			require.Equal(t, test.wantStatus, resp.StatusCode())
			require.Equal(t, test.wantBody, string(resp.Body()))
		})
	}
}

func Test_ipInSubnet(t *testing.T) {
	tests := []struct {
		name          string
		ip            net.IP
		trustedSubnet string
		want          bool
		wantErr       bool
		err           error
	}{
		{
			name:          "ip_in_subnet",
			ip:            net.ParseIP("192.168.1.1"),
			trustedSubnet: "192.168.1.0/24",
			want:          true,
			wantErr:       false,
			err:           nil,
		},
		{
			name:          "ip_not_in_subnet",
			ip:            net.ParseIP("192.168.1.1"),
			trustedSubnet: "192.168.2.0/24",
			want:          false,
			wantErr:       false,
			err:           nil,
		},
		{
			name:          "invalid_subnet",
			ip:            net.ParseIP("192.168.1.1"),
			trustedSubnet: "invalid",
			want:          false,
			wantErr:       true,
			err:           errors.New("invalid CIDR address: invalid"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ok, err := ipInSubnet(tt.ip, tt.trustedSubnet)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, ok)
		})
	}
}
