// Package repository provides data storage abstraction for metrics.
// It supports multiple storage backends including database, file, and in-memory storage.
package repository

import (
	"context"
	"fmt"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository/db"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository/fstorage"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository/mstorage"
	"go.uber.org/zap"
)

// Repository is an interface that defines the methods for storing and retrieving metrics.
type Repository interface {
	// SetGauge sets a gauge metric with the given name and value.
	SetGauge(ctx context.Context, name string, value *float64) error
	// GetGauge retrieves the value of a gauge metric by its name.
	GetGauge(ctx context.Context, name string) (*float64, error)
	// AddCounter increments a counter metric by the given delta.
	AddCounter(ctx context.Context, name string, delta *int64) error
	// GetCounter retrieves the value of a counter metric by its name.
	GetCounter(ctx context.Context, name string) (*int64, error)
	// GetAll returns all available metrics
	GetAll(ctx context.Context) (map[string]string, error)
	// SaveBatch saves a batch of metrics to the repository.
	SaveBatch(ctx context.Context, batch []models.Metrics) error
	// Ping checks the connection to the repository.
	Ping(ctx context.Context) error
	// Close closes the repository.
	Close() error
}

// NewRepository creates a new repository based on the configuration.
func NewRepository(ctx context.Context, config RepositoryConfig, logger *zap.SugaredLogger) (Repository, error) {
	if config.DBConfig.DatabaseDSN != "" {
		logger.Info("Using database storage")
		db, err := db.NewDB(ctx, &config.DBConfig, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create repository: %w", err)
		}
		return db, nil
	} else if config.FSConfig.FPath != "" {
		logger.Info("Using file storage")
		return fstorage.NewFileSaver(ctx, &config.FSConfig, mstorage.NewMemStorage(), logger), nil
	} else {
		logger.Info("Using in-memory storage")
		return mstorage.NewMemStorage(), nil
	}
}
