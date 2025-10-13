package handler

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/audit"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository/mstorage"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func BenchmarkUpdateMetricHandler(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	storage := mstorage.NewMemStorage()

	auditor := audit.NewAuditor(sugaredLogger, "", "")

	handler := NewHandler(storage, "test-key", auditor, sugaredLogger)

	testCases := []struct {
		name        string
		metricType  string
		metricName  string
		metricValue string
	}{
		{"Gauge", "gauge", "test_gauge", "123.45"},
		{"Counter", "counter", "test_counter", "100"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					req := httptest.NewRequest("POST", "/update/"+tc.metricType+"/"+tc.metricName+"/"+tc.metricValue, nil)
					req = req.WithContext(context.Background())

					rctx := chi.NewRouteContext()
					rctx.URLParams.Add("metricType", tc.metricType)
					rctx.URLParams.Add("metricName", tc.metricName)
					rctx.URLParams.Add("metricValue", tc.metricValue)
					req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

					w := httptest.NewRecorder()

					handler.UpdateMetricHandler()(w, req)
				}
			})
		})
	}
}

func BenchmarkGetMetricHandler(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	storage := mstorage.NewMemStorage()

	auditor := audit.NewAuditor(sugaredLogger, "", "")

	handler := NewHandler(storage, "test-key", auditor, sugaredLogger)

	ctx := context.Background()
	testGaugeValue := 123.45
	testCounterValue := int64(100)
	storage.SetGauge(ctx, "test_gauge", &testGaugeValue)
	storage.AddCounter(ctx, "test_counter", &testCounterValue)

	testCases := []struct {
		name       string
		metricType string
		metricName string
	}{
		{"Gauge", "gauge", "test_gauge"},
		{"Counter", "counter", "test_counter"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					req := httptest.NewRequest("GET", "/value/"+tc.metricType+"/"+tc.metricName, nil)
					req = req.WithContext(context.Background())

					rctx := chi.NewRouteContext()
					rctx.URLParams.Add("metricType", tc.metricType)
					rctx.URLParams.Add("metricName", tc.metricName)
					req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

					w := httptest.NewRecorder()

					handler.GetMetricHandler()(w, req)
				}
			})
		})
	}
}

func BenchmarkListMetricsHandler(b *testing.B) {
	// Create a mock logger
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	storage := mstorage.NewMemStorage()

	auditor := audit.NewAuditor(sugaredLogger, "", "")

	handler := NewHandler(storage, "test-key", auditor, sugaredLogger)

	ctx := context.Background()
	for i := 0; i < 100; i++ {
		gaugeValue := float64(i)
		counterValue := int64(i)
		storage.SetGauge(ctx, "gauge_"+string(rune(i)), &gaugeValue)
		storage.AddCounter(ctx, "counter_"+string(rune(i)), &counterValue)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/", nil)
			req = req.WithContext(context.Background())

			w := httptest.NewRecorder()

			handler.ListMetricsHandler()(w, req)
		}
	})
}
