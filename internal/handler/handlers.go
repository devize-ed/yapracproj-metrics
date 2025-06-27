package handler

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	storage "github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	"github.com/go-chi/chi"
)

// handler for updating metrics
func UpdateMetricHandler(ms *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get url parameters
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")
		metricType := chi.URLParam(r, "metricType")

		// handle different metric types, if unkown -> response as http.StatusBadRequest
		switch chi.URLParam(r, "metricType") {
		case models.Counter:
			log.Println("Counter:", metricName, metricValue)
			// convert string value from url and save in the storage
			val, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Incorrect counter value", http.StatusBadRequest)
				return
			}
			ms.AddCounter(metricName, val)
			log.Printf("Counter %s increased by %d\n", metricName, val)

		case models.Gauge:
			log.Println("Gauge", metricName, metricValue)
			// convert string value from url and save in the storage
			val, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Incorrect gauge value", http.StatusBadRequest)
				return
			}
			ms.SetGauge(metricName, val)
			log.Printf("Gauge %s updated to %f\n", metricName, val)

		default:
			// if metric type is unknown, return http.StatusBadRequest
			log.Println("Request invalid metric type: ", metricType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		// write response
		log.Println("Writing response: ", http.StatusOK)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

	}

}

// handler for getting the value of the requested metric
func GetMetricHandler(ms *storage.MemStorage) http.HandlerFunc {
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
			got, ok = ms.GetCounter(metricName)
			if ok {
				val = []byte(strconv.FormatInt(got, 10))
			} else {
				log.Println("Requested metric not found: ", r.URL.Path)
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
		case models.Gauge:
			// get the metric value from the storage, if not found -> response as http.StatusNotFound
			var got float64
			got, ok = ms.GetGauge(metricName)
			if ok {
				val = []byte(strconv.FormatFloat(got, 'f', -1, 64))
			} else {
				log.Println("Requested metric not found: ", r.URL.Path)
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}

		default:
			// if metric type is unknown, return http.StatusBadRequest
			log.Println("Request invalid metric type: ", metricType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		// write response
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(val)
	}
}

// Handler to list all the saved metrics in the storage.
func ListAllHandler(ms *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get the map with all the metrics from the storage
		metrics := ms.ListAll()

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
