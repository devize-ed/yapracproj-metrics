package handler

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	repository "github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	"github.com/go-chi/chi"
)

// Handler wraps the storage.
type Handler struct {
	storage repository.Repository
}

// NewHandler constructs a new Handler with the provided storage.
func NewHandler(r repository.Repository) *Handler {
	return &Handler{
		storage: r,
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
			logger.Log.Debug("Counter:", metricName, metricValue)
			// Convert string value from url and save in the storage.
			val, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Incorrect counter value", http.StatusBadRequest)
				return
			}
			if err := h.storage.AddCounter(r.Context(), metricName, &val); err != nil {
				logger.Log.Error("Failed to add counter:", err)
				http.Error(w, "Failed to add counter", http.StatusInternalServerError)
				return
			}
			logger.Log.Debugf("Counter %s increased by %d\n", metricName, val)

		case models.Gauge:
			logger.Log.Debug("Gauge", metricName, metricValue)
			// Convert string value from url and save in the storage.
			val, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Incorrect gauge value", http.StatusBadRequest)
				return
			}
			if err := h.storage.SetGauge(r.Context(), metricName, &val); err != nil {
				logger.Log.Error("Failed to set gauge:", err)
				http.Error(w, "Failed to set gauge", http.StatusInternalServerError)
				return
			}
			logger.Log.Debugf("Gauge %s updated to %f\n", metricName, val)

		default:
			// If metric type is unknown, return http.StatusBadRequest.
			logger.Log.Debug("Request invalid metric type: ", metricType)
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
				logger.Log.Error("Requested metric not found: ", r.URL.Path)
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
		case models.Gauge:
			// Get the metric value from the storage, if not found -> response as http.StatusNotFound.
			got, err := h.storage.GetGauge(r.Context(), metricName)
			if err == nil {
				val = []byte(strconv.FormatFloat(*got, 'f', -1, 64))
			} else {
				logger.Log.Error("Requested metric not found: ", r.URL.Path)
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}

		default:
			// If metric type is unknown, return http.StatusBadRequest.
			logger.Log.Error("Request invalid metric type: ", metricType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		// Write response
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(val); err != nil {
			logger.Log.Debug("Failed to write response:", err)
		}
	}
}

// ListMetricsHandler handles the listing of all metrics in the storage.
func (h *Handler) ListMetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the map with all the metrics from the storage.
		metrics, err := h.storage.GetAll(r.Context())
		if err != nil {
			logger.Log.Error("Failed to get all metrics:", err)
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
				logger.Log.Debug("Failed to write metric:", err)
			}
		}
	}
}
