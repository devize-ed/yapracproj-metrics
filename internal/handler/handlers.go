package handler

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/go-chi/chi"
)

// Repository defines the storage contract used by handler and mem-storage.
type Repository interface {
	SetGauge(ctx context.Context, name string, value float64)
	GetGauge(ctx context.Context, name string) (float64, bool)
	AddCounter(ctx context.Context, name string, delta int64)
	GetCounter(ctx context.Context, name string) (int64, bool)
	ListAll(ctx context.Context) map[string]string
	SaveBatchToRepo(ctx context.Context, batch []models.Metrics) error
	Ping(ctx context.Context) error
}

// Handler wraps the storage.
type Handler struct {
	storage Repository
}

// NewHandler constructs a new Handler with the provided storage.
func NewHandler(r Repository) *Handler {
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
			h.storage.AddCounter(r.Context(), metricName, val)
			logger.Log.Debugf("Counter %s increased by %d\n", metricName, val)

		case models.Gauge:
			logger.Log.Debug("Gauge", metricName, metricValue)
			// Convert string value from url and save in the storage.
			val, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Incorrect gauge value", http.StatusBadRequest)
				return
			}
			h.storage.SetGauge(r.Context(), metricName, val)
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
		var (
			val []byte
			ok  bool
		)

		// Handle different metric types, if unknown -> response as http.StatusBadRequest.
		switch metricType {
		case models.Counter:
			// Get the metric value from the storage, if not found -> response as http.StatusNotFound.
			var got int64
			got, ok = h.storage.GetCounter(r.Context(), metricName)
			if ok {
				val = []byte(strconv.FormatInt(got, 10))
			} else {
				logger.Log.Error("Requested metric not found: ", r.URL.Path)
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
		case models.Gauge:
			// Get the metric value from the storage, if not found -> response as http.StatusNotFound.
			var got float64
			got, ok = h.storage.GetGauge(r.Context(), metricName)
			if ok {
				val = []byte(strconv.FormatFloat(got, 'f', -1, 64))
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
		w.Write(val)
	}
}

// ListMetricsHandler handles the listing of all metrics in the storage.
func (h *Handler) ListMetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the map with all the metrics from the storage.
		metrics := h.storage.ListAll(r.Context())

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Sort the keys to ensure consistent order.
		keys := make([]string, 0, len(metrics))
		for k := range metrics {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		// Write the metrics to the response.
		for _, k := range keys {
			fmt.Fprintf(w, "%s = %s\n", k, metrics[k])
		}
	}
}
