package handler

import (
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// PingHandler checks the DB connection, returns OK response if the connection is successful.
func (h *Handler) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ping the database.
		h.logger.Debug("Pinging the database")
		if err := h.storage.Ping(r.Context()); err != nil {
			h.logger.Error("Failed to ping the database: %w", err)
			http.Error(w, "Failed to ping the database", http.StatusInternalServerError)
			return
		}
		h.logger.Debug("Database is connected")
		w.WriteHeader(http.StatusOK)
	}
}
