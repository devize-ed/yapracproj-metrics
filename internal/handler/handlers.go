package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// Repository defines the storage contract used by handler and mem-storage.
type Repository interface {
	SetGauge(name string, value float64)
	GetGauge(name string) (float64, bool)
	AddCounter(name string, delta int64)
	GetCounter(name string) (int64, bool)
	ListAll() map[string]string
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
			h.storage.AddCounter(metricName, val)
			logger.Log.Debugf("Counter %s increased by %d\n", metricName, val)

		case models.Gauge:
			logger.Log.Debug("Gauge", metricName, metricValue)
			// Convert string value from url and save in the storage.
			val, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Incorrect gauge value", http.StatusBadRequest)
				return
			}
			h.storage.SetGauge(metricName, val)
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
			got, ok = h.storage.GetCounter(metricName)
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
			got, ok = h.storage.GetGauge(metricName)
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
		metrics := h.storage.ListAll()

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

// UpdateMetricJSONHandler handles the update of a metric based on JSON request body.
func (h *Handler) UpdateMetricJSONHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Decode request body into model struct.
		logger.Log.Debug("Decoding request JSON body")
		body := &models.Metrics{}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(body); err != nil {
			logger.Log.Debug("Cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Log.Debugf("req body: ID = %s, MType = %s, Delta = %v, Value = %v", body.ID, body.MType, body.Delta, body.Value)

		// Get parameters.
		metricName := body.ID
		metricType := body.MType
		// Handle different metric types, if unknown -> response as http.StatusBadRequest.
		switch metricType {
		case models.Counter:
			var metricValue int64
			if body.Delta != nil {
				metricValue = *body.Delta
			} else {
				http.Error(w, "empty counter value", http.StatusNotFound)
				return
			}
			h.storage.AddCounter(metricName, metricValue)
			logger.Log.Debugf("Counter %s increased by %d\n", metricName, metricValue)

		case models.Gauge:
			var metricValue float64
			if body.Value != nil {
				metricValue = *body.Value
			} else {
				http.Error(w, "empty gauge value", http.StatusNotFound)
				return
			}
			h.storage.SetGauge(metricName, metricValue)
			logger.Log.Debugf("Gauge %s updated to %f\n", metricName, metricValue)

		default:
			// If metric type is unknown, return http.StatusBadRequest.
			logger.Log.Debug("Request invalid metric type: ", metricType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		// Write response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}

}

// GetMetricJSONHandler handles the retrieval of a metric based on JSON request body.
func (h *Handler) GetMetricJSONHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Decode request body into model struct.
		logger.Log.Debug("Decoding request JSON body")
		body := &models.Metrics{}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(body); err != nil {
			logger.Log.Debug("Cannot decode request JSON body:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Log.Debugf("req body: ID = %s, MType = %s, Delta = %v, Value = %v", body.ID, body.MType, body.Delta, body.Value)

		// Get parameters.
		metricName := body.ID
		metricType := body.MType

		// Handle different metric types, if unknown -> response as http.StatusBadRequest.
		switch metricType {
		case models.Counter:
			// Get the metric value from the storage, if not found -> response as http.StatusNotFound.
			got, ok := h.storage.GetCounter(metricName)
			if ok {
				body.Delta = &got
			} else {
				logger.Log.Error("Requested metric not found: ", metricName)
				logger.Log.Debugln("Available metrics: ", h.storage.ListAll())
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
		case models.Gauge:
			// Get the metric value from the storage, if not found -> response as http.StatusNotFound.
			got, ok := h.storage.GetGauge(metricName)
			if ok {
				body.Value = &got
			} else {
				logger.Log.Error("Requested metric not found: ", metricName)
				logger.Log.Debugln("Available metrics: ", h.storage.ListAll())
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}

		default:
			// If metric type is unknown, return http.StatusBadRequest.
			logger.Log.Error("Request invalid metric type: ", metricType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		// Write response.
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(body); err != nil {
			logger.Log.Debug("Cannot encode response JSON:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
