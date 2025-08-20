package handler

import (
	"encoding/json"
	"net/http"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"go.uber.org/zap"
)

// UpdateMetricJSONHandler handles the update of a metric based on JSON request body.
func (h *Handler) UpdateMetricJSONHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Decode request body into model struct.
		h.logger.Debug("Decoding request JSON body")
		body := &models.Metrics{}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(body); err != nil {
			h.logger.Debug("Cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.logger.Debugf("req body: ID = %s, MType = %s, Delta = %v, Value = %v", body.ID, body.MType, body.Delta, body.Value)
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

			if err := h.storage.AddCounter(r.Context(), metricName, &metricValue); err != nil {
				h.logger.Error("Failed to add counter:", err)
				http.Error(w, "Failed to add counter", http.StatusInternalServerError)
				return
			}
			h.logger.Debugf("Counter %s increased by %d\n", metricName, metricValue)

		case models.Gauge:
			var metricValue float64
			if body.Value != nil {
				metricValue = *body.Value
			} else {
				http.Error(w, "empty gauge value", http.StatusNotFound)
				return
			}
			if err := h.storage.SetGauge(r.Context(), metricName, &metricValue); err != nil {
				h.logger.Error("Failed to set gauge:", err)
				http.Error(w, "Failed to set gauge", http.StatusInternalServerError)
				return
			}
			h.logger.Debugf("Gauge %s updated to %f\n", metricName, metricValue)

		default:
			// If metric type is unknown, return http.StatusBadRequest.
			h.logger.Debug("Request invalid metric type: ", metricType)
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
		h.logger.Debug("Decoding request JSON body")
		body := &models.Metrics{}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(body); err != nil {
			h.logger.Debug("Cannot decode request JSON body:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.logger.Debugf("req body: ID = %s, MType = %s, Delta = %v, Value = %v", body.ID, body.MType, body.Delta, body.Value)

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
				h.logger.Error("Requested metric not found: ", metricName)
				metrics, err := h.storage.GetAll(r.Context())
				if err == nil {
					h.logger.Debugln("Available metrics: ", metrics)
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
				h.logger.Error("Requested metric not found: ", metricName)
				metrics, err := h.storage.GetAll(r.Context())
				if err == nil {
					h.logger.Debugln("Available metrics: ", metrics)
				}
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}

		default:
			// If metric type is unknown, return http.StatusBadRequest.
			h.logger.Error("Request invalid metric type: ", metricType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		// Write response.
		resp, err := json.Marshal(body)
		if err != nil {
			h.logger.Debug("Cannot encode response JSON:", err)
			http.Error(w, "Cannot encode response JSON", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(resp); err != nil {
			h.logger.Debug("Failed to write response body:", err)
		}
	}
}

// UpdateBatchHandler handles the batch update of metrics based on JSON request body.
func (h *Handler) UpdateBatchHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var metrics []models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			h.logger.Debug("Cannot decode request JSON body", zap.Error(err))
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if err := h.storage.SaveBatch(r.Context(), metrics); err != nil {
			h.logger.Error("failed to save batch", zap.Error(err))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		h.logger.Debug("Saved batch of metrics", zap.Any("batch", metrics))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}
