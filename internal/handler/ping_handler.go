package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// PingHandler checks the DB connection, returns OK response if the connection is successful.
func (h *Handler) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a context with a timeout to avoid hanging if the DB is down.
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		// Ping the database.
		logger.Log.Debug("Pinging the database")
		if err := h.db.PingContext(ctx); err != nil {
			logger.Log.Error("Failed to ping the database: %w", err)
			http.Error(w, "Failed to ping the database", http.StatusInternalServerError)
			return
		}
		logger.Log.Debug("Database is connected")
		w.WriteHeader(http.StatusOK)
	}
}
