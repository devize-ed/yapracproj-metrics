package agent

import (
	"testing"

	"go.uber.org/zap"
)

func BenchmarkCollectRuntimeMetrics(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	storage := NewAgentStorage(sugaredLogger)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			storage.collectRuntimeMetrics()
		}
	})
}

func BenchmarkCollectAdditionalMetrics(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	storage := NewAgentStorage(sugaredLogger)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			storage.collectAdditionalMetrics()
		}
	})
}

func BenchmarkCollectSystemMetrics(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	sugaredLogger := logger.Sugar()

	storage := NewAgentStorage(sugaredLogger)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			storage.collectSystemMetrics()
		}
	})
}
