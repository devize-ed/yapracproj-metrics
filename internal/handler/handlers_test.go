package handler

import (
	"net/http"
	"reflect"
	"testing"

	st "github.com/devize-ed/yapracproj-metrics.git/internal/repository/storage"
)

func TestMakeUpdateHandler(t *testing.T) {
	type args struct {
		storage *st.MemStorage
	}
	tests := []struct {
		name string
		args args
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeUpdateHandler(tt.args.storage); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeUpdateHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	type args struct {
		next http.Handler
	}
	tests := []struct {
		name string
		args args
		want http.Handler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Middleware(tt.args.next); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Middleware() = %v, want %v", got, tt.want)
			}
		})
	}
}
