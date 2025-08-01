package db

import (
	"context"
	"fmt"
	"time"

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
	// Create a context with a timeout to avoid hanging indefinitely
	context, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Parse the DSN and create a new connection pool with tracing enabled
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}

	if err = runMigrations(context, dsn); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Set the connection pool configuration
	poolCfg.ConnConfig.Tracer = &queryTracer{}
	pool, err := pgxpool.NewWithConfig(context, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	// Ping the database to ensure the connection is established
	if err := pool.Ping(context); err != nil {
		return nil, fmt.Errorf("failed to ping the DB: %w", err)
	}
	return pool, nil
}

// Save writes the metrics to the database.
func (db *DB) Save(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	logger.Log.Debug("Saving metrics to the database")
	// Begin a transaction
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Prepare and execute the SQL statements to insert or update gauges
	for id, val := range gauges {
		if _, err := tx.Exec(ctx, "INSERT INTO gauges (id, value) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value", id, val); err != nil {
			return fmt.Errorf("failed to insert gauge %s: %w", id, err)
		}
	}

	// Prepare and execute the SQL statements to insert or update counters
	for id, val := range counters {
		if _, err := tx.Exec(ctx, "INSERT INTO counters (id, value) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value", id, val); err != nil {
			return fmt.Errorf("failed to insert counter %s: %w", id, err)
		}
	}

	// Commit the transaction
	return tx.Commit(ctx)
}

// Load reads the metrics from the database.
func (db *DB) Load(ctx context.Context) (map[string]float64, map[string]int64, error) {
	logger.Log.Debug("Loading metrics from the database")
	// create temporary maps to hold the loaded metrics
	tmp := struct {
		Gauge   map[string]float64 `json:"gauge"`
		Counter map[string]int64   `json:"counter"`
	}{}

	rows, err := db.pool.Query(ctx, "SELECT id, value FROM gauges")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query gauges: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id    string
			value float64
		)
		if err := rows.Scan(&id, &value); err != nil {
			return nil, nil, fmt.Errorf("failed to scan gauge row: %w", err)
		}
		tmp.Gauge[id] = value
	}

	rows, err = db.pool.Query(ctx, "SELECT id, value FROM counters")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query gauges: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id    string
			delta int64
		)
		if err := rows.Scan(&id, &delta); err != nil {
			return nil, nil, fmt.Errorf("failed to scan gauge row: %w", err)
		}
		tmp.Counter[id] = delta
	}

	logger.Log.Debugf("metrics restored from the database: %d gauges, %d counters", len(tmp.Gauge), len(tmp.Counter))
	return tmp.Gauge, tmp.Counter, nil
}

func (db *DB) Ping(ctx context.Context) error {
	logger.Log.Debug("Pinging the database")
	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping the database: %w", err)
	}
	logger.Log.Debug("Database is connected")
	return nil
}

// Close closes the database connection pool.
func (db *DB) Close() error {
	db.pool.Close()
	return nil
}
