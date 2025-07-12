package handler

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	"github.com/go-chi/chi"
)

// Repository interface for interacting with the storage.
type Repository interface {
	SetGauge(name string, value float64)
	GetGauge(name string) (float64, bool)
	AddCounter(name string, delta int64)
	GetCounter(name string) (int64, bool)
	ListAll() map[string]string
}

type Handler struct {
	storage Repository
}

func NewHandler(r Repository) *Handler {
	return &Handler{
		storage: r,
	}
}

// handler for update the value of the requested metric
func (h *Handler) UpdateMetricHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get url parameters
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")
		metricType := chi.URLParam(r, "metricType")

		// handle different metric types, if unkown -> response as http.StatusBadRequest
		switch chi.URLParam(r, "metricType") {
		case models.Counter:
			logger.Log.Debug("Counter:", metricName, metricValue)
			// convert string value from url and save in the storage
			val, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Incorrect counter value", http.StatusBadRequest)
				return
			}
			h.storage.AddCounter(metricName, val)
			logger.Log.Debugf("Counter %s increased by %d\n", metricName, val)

		case models.Gauge:
			logger.Log.Debug("Gauge", metricName, metricValue)
			// convert string value from url and save in the storage
			val, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Incorrect gauge value", http.StatusBadRequest)
				return
			}
			h.storage.SetGauge(metricName, val)
			logger.Log.Debugf("Gauge %s updated to %f\n", metricName, val)

		default:
			// if metric type is unknown, return http.StatusBadRequest
			logger.Log.Debug("Request invalid metric type: ", metricType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		// write response
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}

}

// handler for getting the value of the requested metric
func (h *Handler) GetMetricHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get url parameters
		metricName := chi.URLParam(r, "metricName")
		metricType := chi.URLParam(r, "metricType")

		// init vats to find and convert the metric value
		var (
			val []byte
			ok  bool
		)

		// handle different metric types, if unkown -> response as http.StatusBadRequest
		switch metricType {
		case models.Counter:
			// get the metric value from the storage, if not found -> response as http.StatusNotFound
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
			// get the metric value from the storage, if not found -> response as http.StatusNotFound
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
			// if metric type is unknown, return http.StatusBadRequest
			logger.Log.Error("Request invalid metric type: ", metricType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		// write response
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write(val)
	}
}

// Handler to list all the saved metrics in the storage.
func (h *Handler) ListMetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get the map with all the metrics from the storage
		metrics := h.storage.ListAll()

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		// sort the keys
		keys := make([]string, 0, len(metrics))
		for k := range metrics {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		// write the metrics to the response
		for _, k := range keys {
			fmt.Fprintf(w, "%s = %s\n", k, metrics[k])
		}
	}
}
