package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/audit"
	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/handler"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository/mstorage"
	"go.uber.org/zap"
)

func BenchmarkServerCreation(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	storage := mstorage.NewMemStorage()
	auditor := audit.NewAuditor(sugaredLogger, "", "")
	handler := handler.NewHandler(storage, "test-key", auditor, sugaredLogger)

	cfg := config.ServerConfig{
		Connection: config.ServerConn{
			Host: "localhost:8080",
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = NewServer(cfg, handler, sugaredLogger)
		}
	})
}

func BenchmarkServerShutdown(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	storage := mstorage.NewMemStorage()
	auditor := audit.NewAuditor(sugaredLogger, "", "")
	handler := handler.NewHandler(storage, "test-key", auditor, sugaredLogger)

	cfg := config.ServerConfig{
		Connection: config.ServerConn{
			Host: "localhost:8080",
		},
	}

	server := NewServer(cfg, handler, sugaredLogger)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
			_ = server.shutdown(ctx)
			cancel()
		}
	})
}

func BenchmarkServerRequestHandling(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	storage := mstorage.NewMemStorage()
	auditor := audit.NewAuditor(sugaredLogger, "", "")
	handler := handler.NewHandler(storage, "test-key", auditor, sugaredLogger)

	cfg := config.ServerConfig{
		Connection: config.ServerConn{
			Host: "localhost:8080",
		},
	}

	server := NewServer(cfg, handler, sugaredLogger)

	testServer := httptest.NewServer(server.Handler)
	defer testServer.Close()

	testCases := []struct {
		name    string
		method  string
		path    string
		headers map[string]string
	}{
		{"GET_Ping", "GET", "/ping", nil},
		{"GET_Value", "GET", "/value/gauge/test_gauge", nil},
		{"POST_Update", "POST", "/update/gauge/test_gauge/123.45", nil},
		{"GET_All", "GET", "/", nil},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					req, _ := http.NewRequest(tc.method, testServer.URL+tc.path, nil)
					if tc.headers != nil {
						for k, v := range tc.headers {
							req.Header.Set(k, v)
						}
					}

					client := &http.Client{Timeout: 1 * time.Second}
					resp, err := client.Do(req)
					if err != nil {
						logger.Error("error making HTTP request", zap.Error(err))
						b.Fail()
						return
					}
					_ = resp.Body.Close()
				}
			})
		})
	}
}

func BenchmarkServerRouterCreation(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	storage := mstorage.NewMemStorage()
	auditor := audit.NewAuditor(sugaredLogger, "", "")
	handler := handler.NewHandler(storage, "test-key", auditor, sugaredLogger)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = handler.NewRouter()
		}
	})
}
