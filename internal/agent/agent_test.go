package agent

import (
	"net/http"
	"testing"
)

func TestSendMetric(t *testing.T) {
	type args struct {
		client *http.Client
		metric string
		value  string
		host   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendMetric(tt.args.client, tt.args.metric, tt.args.value, tt.args.host); (err != nil) != tt.wantErr {
				t.Errorf("SendMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
