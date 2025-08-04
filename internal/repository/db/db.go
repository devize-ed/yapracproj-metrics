package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/devize-ed/yapracproj-metrics.git/migrations"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents a database connection pool.
type DB struct {
	pool *pgxpool.Pool
}

// NewDB provides the new data base connection with the provided configuration.
func NewDB(ctx context.Context, dsn string) (*DB, error) {
	logger.Log.Debugf("Connecting to database with DSN: %s", dsn)

	// Run migrations before establishing the connection
	if err := migrations.RunMigrations(dsn, true); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}

	// Initialize a new connection pool with the provided DSN
	pool, err := initPool(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise a connection pool: %w", err)
	}

	logger.Log.Debug("Database connection established successfully")
	return &DB{pool: pool}, nil
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
		if _, err := tx.Exec(ctx, `INSERT INTO gauges (id, value) 
				VALUES ($1, $2) 
				ON CONFLICT (id) DO UPDATE 
				SET value = EXCLUDED.value 
			`, id, val); err != nil {
			return fmt.Errorf("failed to insert gauge %s: %w", id, err)
		}
	}

	// Prepare and execute the SQL statements to insert or update counters
	for id, val := range counters {
		if _, err := tx.Exec(ctx, `
				INSERT INTO counters (id, delta) 
				VALUES ($1, $2) 
				ON CONFLICT (id) DO UPDATE 
				SET delta = counters.delta + EXCLUDED.delta
			`, id, val); err != nil {
			return fmt.Errorf("failed to insert counter %s: %w", id, err)
		}
	}

	// Commit the transaction
	if err := commitWithRetries(ctx, tx); err != nil {
		return fmt.Errorf("commit error: %w", err)
	}
	return nil
}

// SaveBatch saves a batch of metrics to the database.
func (db *DB) SaveBatch(ctx context.Context, metrics []models.Metrics) error {
	// Begin a transaction
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Prepare a batch of SQL statements to insert or update metrics
	batch := &pgx.Batch{}
	for _, m := range metrics {
		switch m.MType {
		case models.Gauge:
			batch.Queue(`
                INSERT INTO gauges(id,value)
                VALUES ($1,$2)
                ON CONFLICT(id) DO UPDATE
                SET value = EXCLUDED.value
            `, m.ID, m.Value)
		case models.Counter:
			batch.Queue(`
                INSERT INTO counters(id,delta)
                VALUES ($1,$2)
                ON CONFLICT(id) DO UPDATE
                SET delta = counters.delta + EXCLUDED.delta
            `, m.ID, m.Delta)
		}
	}

	// Send the batch to the database
	br := tx.SendBatch(ctx, batch)

	// Check for errors in the batch execution
	if err := br.Close(); err != nil {
		return fmt.Errorf("batch close: %w", err)
	}

	// Commit the transaction
	if err := commitWithRetries(ctx, tx); err != nil {
		return fmt.Errorf("commit error: %w", err)
	}
	return nil
}

// Load reads the metrics from the database.
func (db *DB) Load(ctx context.Context) (map[string]float64, map[string]int64, error) {
	tx, err := db.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return map[string]float64{}, map[string]int64{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	logger.Log.Debug("Loading metrics from the database")
	// create temporary maps to hold the loaded metrics
	gauge := map[string]float64{}
	counter := map[string]int64{}

	rows, err := tx.Query(ctx, "SELECT id, value FROM gauges")
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
		gauge[id] = value
	}

	rows, err = tx.Query(ctx, "SELECT id, delta FROM counters")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query counters: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id    string
			delta int64
		)
		if err := rows.Scan(&id, &delta); err != nil {
			return nil, nil, fmt.Errorf("failed to scan counters row: %w", err)
		}
		counter[id] = delta
	}

	// Commit the transaction
	if err := commitWithRetries(ctx, tx); err != nil {
		return map[string]float64{}, map[string]int64{}, fmt.Errorf("commit error: %w", err)
	}
	logger.Log.Debugf("metrics restored from the database: %d gauges, %d counters", len(gauge), len(counter))
	return gauge, counter, nil
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

func commitWithRetries(ctx context.Context, tx pgx.Tx) error {
	// Define backoff durations for retries
	backoffs := []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

	for i := 0; ; i++ {
		// attempt to commit the transaction
		if err := tx.Commit(ctx); err != nil {
			// if the error is not retriable, return it
			if !isErrorRetriable(err) {
				return fmt.Errorf("commit (attempt %d): %w", i+1, err)
			}
			// if we have exhausted all retries, return the error
			if i == len(backoffs) {
				return fmt.Errorf("commit (attempt %d): %w", i+1, err)
			}
			// if the error is retriable, wait for the backoff duration and retry
			select {
			case <-time.After(backoffs[i]):
				continue // retry the commit
			case <-ctx.Done():
				return ctx.Err() // return context error if the context is done
			}
		}
		// if the commit was successful, return nil
		return nil
	}
}

// isErrorRetriable checks for specific PostgreSQL error codes that indicate retriable errors (connection issues).
func isErrorRetriable(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && (pgErr.Code == pgerrcode.SerializationFailure ||
		pgErr.Code == pgerrcode.LockNotAvailable ||
		pgErr.Code == pgerrcode.QueryCanceled) {
		return true
	}
	return false
}
