package db

import (
	"context"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/jackc/pgx/v5"
)

// queryTracer implements the pgx.Tracer interface to log query execution details.
type queryTracer struct{}

// TraceQueryStart logs the start of a query execution.
func (t *queryTracer) TraceQueryStart(
	ctx context.Context,
	_ *pgx.Conn,
	data pgx.TraceQueryStartData,
) context.Context {
	logger.Log.Debugf("Running query %s (%v)", data.SQL, data.Args)
	return ctx
}

// TraceQueryEnd logs the end of a query execution.
func (t *queryTracer) TraceQueryEnd(_ context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	logger.Log.Debugf("%v", data.CommandTag)
}
