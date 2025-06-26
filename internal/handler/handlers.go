package handler

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	st "github.com/devize-ed/yapracproj-metrics.git/internal/repository/storage"
	"github.com/go-chi/chi"
)

func UpdateMetricHandler(storage *st.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")
		metricType := chi.URLParam(r, "metricType")

		switch chi.URLParam(r, "metricType") {
		case models.Counter:
			log.Println("Counter:", metricName, metricValue)
			val, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Incorrect counter value", http.StatusBadRequest)
				return
			}
			storage.AddCounter(metricName, val)
			log.Printf("Counter %s increased by %d\n", metricName, val)

		case models.Gauge:
			log.Println("Gauge", metricName, metricValue)
			val, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Incorrect gauge value", http.StatusBadRequest)
				return
			}
			storage.SetGauge(metricName, val)
			log.Printf("Gauge %s updated to %f\n", metricName, val)

		default:
			log.Println("Request invalid metric type: ", metricType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		log.Println("Writing response: ", http.StatusOK)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

	}

}

func GetMetricHandler(storage *st.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricName := chi.URLParam(r, "metricName")
		metricType := chi.URLParam(r, "metricType")
		log.Println("GetMetricHandler called with metricName:", metricName, "and metricType:", metricType)

		var (
			val []byte
			ok  bool
		)

		switch metricType {
		case models.Counter:
			var got int64
			got, ok = storage.GetCounter(metricName)
			if ok {
				val = []byte(strconv.FormatInt(got, 10))
			} else {
				log.Println("Requested metric not found: ", r.URL.Path)
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
		case models.Gauge:
			var got float64
			got, ok = storage.GetGauge(metricName)
			if ok {
				val = []byte(strconv.FormatFloat(got, 'f', -1, 64))
			} else {
				log.Println("Requested metric not found: ", r.URL.Path)
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}

		default:
			log.Println("Request invalid metric type: ", metricType)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(val)
	}
}

func ListAllHandler(storage *st.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := storage.ListAll()

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		keys := make([]string, 0, len(metrics))
		for k := range metrics {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(w, "%s = %s\n", k, metrics[k])
		}
	}
}
