package repository

import (
	"context"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
)

// StubStorage is a mock implementation of the Storage interface for the case of usinf internal server memory storage.
type StubStorage struct{}

// NewStubRepository creates a new instance of StubStorage for testing purposes.
func NewStubStorage() *StubStorage {
	return &StubStorage{}
}

// Save and Load methods are no-ops for the stub repository.
func (s *StubStorage) Save(ctx context.Context, gauge map[string]float64, counter map[string]int64) error {
	return nil // No-op for stub
}

// Save and Load methods are no-ops for the stub repository.
func (s *StubStorage) SaveBatch(ctx context.Context, metrics []models.Metrics) error {
	return nil // No-op for stub
}

// Load returns empty maps for the stub repository.
func (s *StubStorage) Load(ctx context.Context) (map[string]float64, map[string]int64, error) {
	return map[string]float64{}, map[string]int64{}, nil // No-op for stub
}
