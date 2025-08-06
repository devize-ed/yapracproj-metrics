package repository

import (
	"context"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository/db"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository/fstorage"
	mstorage "github.com/devize-ed/yapracproj-metrics.git/internal/repository/mstorage"
)

type Repository interface {
	// SetGauge sets a gauge metric with the given name and value.
	SetGauge(ctx context.Context, name string, value *float64) error
	// GetGauge retrieves the value of a gauge metric by its name.
	GetGauge(ctx context.Context, name string) (*float64, error)
	// AddCounter increments a counter metric by the given delta.
	AddCounter(ctx context.Context, name string, delta *int64) error
	// GetCounter retrieves the value of a counter metric by its name.
	GetCounter(ctx context.Context, name string) (*int64, error)
	// ListAll returns all available metrics
	GetAll(ctx context.Context) (map[string]string, error)
	// SaveBatchToRepo saves a batch of metrics to the repository.
	SaveBatch(ctx context.Context, batch []models.Metrics) error
	// Ping checks the connection to the repository.
	Ping(ctx context.Context) error
	// Close closes the repository.
	Close() error
}

func NewRepository(ctx context.Context, config RepositoryConfig) Repository {
	if config.DBConfig.DatabaseDSN != "" {
		logger.Log.Info("Using database storage")
		db, err := db.NewDB(ctx, &config.DBConfig)
		if err != nil {
			logger.Log.Errorf("failed to create repository: %v", err)
		}
		return db
	} else if config.FileConfig.FPath != "" {
		logger.Log.Info("Using file storage")
		return fstorage.NewFileSaver(ctx, &config.FileConfig, mstorage.NewMemStorage())
	} else {
		logger.Log.Info("Using in-memory storage")
		return mstorage.NewMemStorage()
	}
}
