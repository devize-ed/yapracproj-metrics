package db

import (
	"context"
	"fmt"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents a database connection pool.
type DB struct {
	pool *pgxpool.Pool
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
		pool: pool,
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
	db.pool.Close()
	return nil
}

// func initDBConnection(cfg config.ServerConfig) (*sql.DB, error) {
// 	// Check if the database DSN is provided
// 	if cfg.DatabaseDSN == "" {
// 		logger.Log.Debug("No database DSN provided, skipping database initialization")
// 		return nil, nil
// 	}

// 	logger.Log.Debugf("Connecting to database with DSN: %s", cfg.DatabaseDSN)
// 	// Open a database connection using loaded configuration
// 	db, err := sql.Open("pgx", cfg.DatabaseDSN)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open sql driver: %w", err)
// 	}

// 	// Ping the database to ensure the connection is established
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()
// 	if err := db.PingContext(ctx); err != nil {
// 		return nil, fmt.Errorf("failed to ping db: %w", err)
// 	}
// 	logger.Log.Debug("Database connection established successfully")
// 	return db, nil
// }
