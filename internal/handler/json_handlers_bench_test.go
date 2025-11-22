package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/audit"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository/mstorage"
	"go.uber.org/zap"
)

func BenchmarkUpdateMetricJSONHandler(b *testing.B) {
	testCases := []struct {
		name        string
		metricType  string
		metricName  string
		metricValue interface{}
	}{
		{"Gauge", models.Gauge, "test_gauge", 123.45},
		{"Counter", models.Counter, "test_counter", int64(100)},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			var body models.Metrics
			if tc.metricType == models.Gauge {
				value := tc.metricValue.(float64)
				body = models.Metrics{
					ID:    tc.metricName,
					MType: tc.metricType,
					Value: &value,
				}
			} else {
				delta := tc.metricValue.(int64)
				body = models.Metrics{
					ID:    tc.metricName,
					MType: tc.metricType,
					Delta: &delta,
				}
			}

			jsonBody, _ := json.Marshal(body)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				logger, _ := zap.NewDevelopment()
				sugaredLogger := logger.Sugar()
				storage := mstorage.NewMemStorage()
				auditor := audit.NewAuditor(sugaredLogger, "", "")
				handler := NewHandler(storage, "test-key", auditor, sugaredLogger)

				for pb.Next() {
					req := httptest.NewRequest("POST", "/update", bytes.NewReader(jsonBody))
					req.Header.Set("Content-Type", "application/json")
					req = req.WithContext(context.Background())

					w := httptest.NewRecorder()
					handler.UpdateMetricJSONHandler()(w, req)
				}
			})
		})
	}
}

func BenchmarkGetMetricJSONHandler(b *testing.B) {
	testCases := []struct {
		name       string
		metricType string
		metricName string
	}{
		{"Gauge", models.Gauge, "test_gauge"},
		{"Counter", models.Counter, "test_counter"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Create JSON request body
			body := models.Metrics{
				ID:    tc.metricName,
				MType: tc.metricType,
			}

			jsonBody, _ := json.Marshal(body)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				// Create separate instances for each goroutine to avoid race conditions
				logger, _ := zap.NewDevelopment()
				sugaredLogger := logger.Sugar()
				storage := mstorage.NewMemStorage()
				auditor := audit.NewAuditor(sugaredLogger, "", "")
				handler := NewHandler(storage, "test-key", auditor, sugaredLogger)

				// Pre-populate storage with test data
				ctx := context.Background()
				testGaugeValue := 123.45
				testCounterValue := int64(100)
				if err := storage.SetGauge(ctx, "test_gauge", &testGaugeValue); err != nil {
					b.Fatal(err)
				}
				if err := storage.AddCounter(ctx, "test_counter", &testCounterValue); err != nil {
					b.Fatal(err)
				}

				for pb.Next() {
					req := httptest.NewRequest("POST", "/value", bytes.NewReader(jsonBody))
					req.Header.Set("Content-Type", "application/json")
					req = req.WithContext(context.Background())

					w := httptest.NewRecorder()
					handler.GetMetricJSONHandler()(w, req)
				}
			})
		})
	}
}

func BenchmarkUpdateBatchHandler(b *testing.B) {
	metrics := make([]models.Metrics, 10)
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			value := float64(i)
			metrics[i] = models.Metrics{
				ID:    "gauge_" + string(rune(i)),
				MType: models.Gauge,
				Value: &value,
			}
		} else {
			delta := int64(i)
			metrics[i] = models.Metrics{
				ID:    "counter_" + string(rune(i)),
				MType: models.Counter,
				Delta: &delta,
			}
		}
	}

	jsonBody, _ := json.Marshal(metrics)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		logger, _ := zap.NewDevelopment()
		sugaredLogger := logger.Sugar()
		storage := mstorage.NewMemStorage()
		auditor := audit.NewAuditor(sugaredLogger, "", "")
		handler := NewHandler(storage, "test-key", auditor, sugaredLogger)

		for pb.Next() {
			req := httptest.NewRequest("POST", "/updates", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.Background())

			w := httptest.NewRecorder()
			handler.UpdateBatchHandler()(w, req)
		}
	})
}

func BenchmarkMetricsToStrings(b *testing.B) {
	metrics := make([]models.Metrics, 100)
	for i := 0; i < 100; i++ {
		metrics[i] = models.Metrics{
			ID:    "metric_" + string(rune(i)),
			MType: models.Gauge,
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = metricsToStrings(metrics)
		}
	})
}
