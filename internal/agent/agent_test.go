package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestSendMetric(t *testing.T) {
	type args struct {
		metric string
		value  string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantCode int
	}{
		{
			name: "Send metric",
			args: args{
				metric: "testMetric",
				value:  "123",
			},
			wantErr:  false,
			wantCode: http.StatusOK,
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	host := strings.TrimPrefix(srv.URL, "http://")
	client := resty.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendMetric(client, tt.args.metric, tt.args.value, host); (err != nil) != tt.wantErr {
				t.Errorf("SendMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		assert.Equal(t, tt.wantCode, http.StatusOK, "Expected status code to be %d", tt.wantCode)
	}
}
