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
			logger.Log.Debug("Cannot encode response JSON: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
