package db

import (
	"context"
	"fmt"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents a database connection pool.
type DB struct {
	Pool *pgxpool.Pool
}

// NewDB provides the new data base connection with the provided configuration.
func NewDB(ctx context.Context, dsn string) (*DB, error) {
	logger.Log.Debugf("Connecting to database with DSN: %s", dsn)
	pool, err := initPool(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	logger.Log.Debug("Database connection established successfully")
	return &DB{
		Pool: pool,
	}, nil
}

// initPool initializes a new connection pool.
func initPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	// Parse the DSN and create a new connection pool with tracing enabled
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}

	// Set the connection pool configuration
	poolCfg.ConnConfig.Tracer = &queryTracer{}
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	// Ping the database to ensure the connection is established
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping the DB: %w", err)
	}
	return pool, nil
}

// Close closes the database connection pool.
func (db *DB) Close() error {
	db.Pool.Close()
	return nil
}
