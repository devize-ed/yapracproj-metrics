package db

import (
	"context"
	"fmt"

	"github.com/devize-ed/yapracproj-metrics.git/internal/config/db"
)

// FileSaver is a struct that implements the Repository interface for saving metrics to a file.
type DB struct {
	db *db.DB
	Close()	error
}

// NewFileSaver constructs a new FileSaver with the provided file name.
func NewDBRepository(ctx context.Context, DSN string) (*DB, error) {
	db, err := db.NewDB(ctx, DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a DB store: %w", err)
	}
	return &DB{
		db: db,
	}, nil
}

// Save writes the metrics to the specified file in JSON format.
func (d *DB) Save(gauge map[string]float64, counter map[string]int64) error {
	println(gauge, counter)
	return nil
}

// Load reads the metrics from the specified file and restores them to the storage.
func (d *DB) Load() (map[string]float64, map[string]int64, error) {
	return nil, nil, nil
}
