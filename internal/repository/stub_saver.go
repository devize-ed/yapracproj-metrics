package repository

import "context"

// StubRepository is a mock implementation of the Repository interface for the case of usinf internal server memory storage.
type StubRepository struct{}

// NewStubRepository creates a new instance of StubRepository for testing purposes.
func NewStubRepository() *StubRepository {
	return &StubRepository{}
}

// Save and Load methods are no-ops for the stub repository.
func (s *StubRepository) Save(ctx context.Context, gauge map[string]float64, counter map[string]int64) error {
	return nil // No-op for stub
}

// Load returns empty maps for the stub repository.
func (s *StubRepository) Load(ctx context.Context) (map[string]float64, map[string]int64, error) {
	return nil, nil, nil // No-op for stub
}
