package handler

import (
	"net/http"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// PingHandler checks the DB connection, returns OK response if the connection is successful.
func (h *Handler) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ping the database.
		logger.Log.Debug("Pinging the database")
		if err := h.storage.Repository.Ping(r.Context()); err != nil {
			logger.Log.Error("Failed to ping the database: %w", err)
			http.Error(w, "Failed to ping the database", http.StatusInternalServerError)
			return
		}
		logger.Log.Debug("Database is connected")
		w.WriteHeader(http.StatusOK)
	}
}
