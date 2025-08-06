package handler

import (
	"encoding/json"
	"net/http"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"go.uber.org/zap"
)

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

			h.storage.AddCounter(r.Context(), metricName, &metricValue)
			logger.Log.Debugf("Counter %s increased by %d\n", metricName, metricValue)

		case models.Gauge:
			var metricValue float64
			if body.Value != nil {
				metricValue = *body.Value
			} else {
				http.Error(w, "empty gauge value", http.StatusNotFound)
				return
			}
			h.storage.SetGauge(r.Context(), metricName, &metricValue)
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
			got, err := h.storage.GetCounter(r.Context(), metricName)
			if err == nil {
				body.Delta = got
			} else {
				logger.Log.Error("Requested metric not found: ", metricName)
				metrics, err := h.storage.GetAll(r.Context())
				if err == nil {
					logger.Log.Debugln("Available metrics: ", metrics)
				}
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
		case models.Gauge:
			// Get the metric value from the storage, if not found -> response as http.StatusNotFound.
			got, err := h.storage.GetGauge(r.Context(), metricName)
			if err == nil {
				body.Value = got
			} else {
				logger.Log.Error("Requested metric not found: ", metricName)
				metrics, err := h.storage.GetAll(r.Context())
				if err == nil {
					logger.Log.Debugln("Available metrics: ", metrics)
				}
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
		resp, err := json.Marshal(body)
		if err != nil {
			logger.Log.Debug("Cannot encode response JSON:", err)
			http.Error(w, "Cannot encode response JSON", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(resp); err != nil {
			logger.Log.Debug("Failed to write response body:", err)
		}
	}
}

// UpdateBatchHandler handles the batch update of metrics based on JSON request body.
func (h *Handler) UpdateBatchHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var metrics []models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			logger.Log.Debug("Cannot decode request JSON body", zap.Error(err))
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if err := h.storage.SaveBatch(r.Context(), metrics); err != nil {
			logger.Log.Error("failed to save batch", zap.Error(err))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		logger.Log.Debug("Saved batch of metrics", zap.Any("batch", metrics))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}
