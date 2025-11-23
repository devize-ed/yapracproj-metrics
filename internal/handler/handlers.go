// Package handler provides HTTP handlers for metric operations.
// It handles metric updates, retrieval, and listing with proper error handling.
package handler

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/devize-ed/yapracproj-metrics.git/internal/audit"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// Handler wraps the storage.
type Handler struct {
	storage       repository.Repository // storage for metrics
	hashKey       string                // key for hashing requests
	auditor       *audit.Auditor        // audito servic for logging changes of metrics
	trustedSubnet string                // trusted subnet for IP filtering
	logger        *zap.SugaredLogger
}

// NewHandler constructs a new Handler with the provided storage.
func NewHandler(r repository.Repository, key string, auditor *audit.Auditor, trustedSubnet string, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		storage:       r,
		hashKey:       key,
		auditor:       auditor,
		trustedSubnet: trustedSubnet,
		logger:        logger,
	}
}

// UpdateMetricHandler handles the update of a metric based on URL parameters.
func (h *Handler) UpdateMetricHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get URL parameters.
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")
		metricType := chi.URLParam(r, "metricType")

		// Handle different metric types, if unknown -> response as http.StatusBadRequest.
		switch chi.URLParam(r, "metricType") {
		case models.Counter:
			h.logger.Debug("Counter:", metricName, metricValue)
			// Convert string value from url and save in the storage.
			val, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Incorrect counter value", http.StatusBadRequest)
				return
			}
			if err := h.storage.AddCounter(r.Context(), metricName, &val); err != nil {
				h.logger.Error("Failed to add counter:", err)
				http.Error(w, "Failed to add counter", http.StatusInternalServerError)
				return
			}
			h.logger.Debugf("Counter %s increased by %d\n", metricName, val)

		case models.Gauge:
			h.logger.Debug("Gauge", metricName, metricValue)
			// Convert string value from url and save in the storage.
			val, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Incorrect gauge value", http.StatusBadRequest)
				return
			}
			if err := h.storage.SetGauge(r.Context(), metricName, &val); err != nil {
				h.logger.Error("Failed to set gauge:", err)
				http.Error(w, "Failed to set gauge", http.StatusInternalServerError)
				return
			}
			h.logger.Debugf("Gauge %s updated to %f\n", metricName, val)

		default:
			// If metric type is unknown, return http.StatusBadRequest.
			h.logger.Debug("Request invalid metric type: ", metricType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		// Write response.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

}

// GetMetricHandler handles the retrieval of a metric based on URL parameters.
func (h *Handler) GetMetricHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get URL parameters.
		metricName := chi.URLParam(r, "metricName")
		metricType := chi.URLParam(r, "metricType")

		// Initialize variables to find and convert the metric value
		var val []byte

		// Handle different metric types, if unknown -> response as http.StatusBadRequest.
		switch metricType {
		case models.Counter:
			// Get the metric value from the storage, if not found -> response as http.StatusNotFound.
			got, err := h.storage.GetCounter(r.Context(), metricName)
			if err == nil {
				val = []byte(strconv.FormatInt(*got, 10))
			} else {
				h.logger.Error("Requested metric not found: ", r.URL.Path)
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
		case models.Gauge:
			// Get the metric value from the storage, if not found -> response as http.StatusNotFound.
			got, err := h.storage.GetGauge(r.Context(), metricName)
			if err == nil {
				val = []byte(strconv.FormatFloat(*got, 'f', -1, 64))
			} else {
				h.logger.Error("Requested metric not found: ", r.URL.Path)
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}

		default:
			// If metric type is unknown, return http.StatusBadRequest.
			h.logger.Error("Request invalid metric type: ", metricType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		// Write response
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(val); err != nil {
			h.logger.Debug("Failed to write response:", err)
		}
	}
}

// ListMetricsHandler handles the listing of all metrics in the storage.
func (h *Handler) ListMetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the map with all the metrics from the storage.
		metrics, err := h.storage.GetAll(r.Context())
		if err != nil {
			h.logger.Error("Failed to get all metrics:", err)
			http.Error(w, "Failed to get all metrics", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Sort the keys to ensure consistent order.
		keys := make([]string, 0, len(metrics))
		for k := range metrics {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		// Write the metrics to the response.
		for _, k := range keys {
			if _, err := fmt.Fprintf(w, "%s = %s\n", k, metrics[k]); err != nil {
				h.logger.Debug("Failed to write metric:", err)
			}
		}
	}
}
