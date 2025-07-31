package db

import (
	"context"

	"github.com/devize-ed/yapracproj-metrics.git/internal/config/db"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
)

// FileSaver is a struct that implements the Repository interface for saving metrics to a file.
type DB struct {
	db *db.DB
}

// NewFileSaver constructs a new FileSaver with the provided file name.
func NewDB(ctx context.Context, DSN string) *DB {
	if db := db.NewDB(ctx, DSN); db == nil {
		logger.Log.Fatal("failed to initialize database connection")
		return DB{
			db: db.NewDB(ctx, DSN),
		}
	}
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
