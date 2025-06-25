package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	models "github.com/devize-ed/yapracproj-metrics.git/internal/model"
	st "github.com/devize-ed/yapracproj-metrics.git/internal/repository/storage"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received update request: ", r.URL.Path)
		if r.Method != http.MethodPost {
			fmt.Println("Request method not allowed: ", r.URL.Path)
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func MakeUpdateHandler(storage *st.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")

		// 0 = ""
		// 1 = "update"
		// 2 = "<ТИП_МЕТРИКИ>"
		// 3 = "<ИМЯ_МЕТРИКИ>"
		// 4 = "<ЗНАЧЕНИЕ_МЕТРИКИ>"

		if len(parts) != 5 || parts[3] == "" || parts[4] == "" {
			fmt.Println("Invalid request or empty metric name/value: ", r.URL.Path)
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		metricType, metricName, metricValue := parts[2], parts[3], parts[4]

		switch metricType {
		case models.Counter:
			fmt.Println("Counter:", metricName, metricValue)
			val, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Incorrect counter value", http.StatusBadRequest)
				return
			}
			storage.AddCounter(metricName, val)
			fmt.Printf("Counter %s increased by %d\n", metricName, val)

		case models.Gauge:
			fmt.Println("Gauge", metricName, metricValue)
			val, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Incorrect gauge value", http.StatusBadRequest)
				return
			}
			storage.SetGauge(metricName, val)
			fmt.Printf("Gauge %s updated to %f\n", metricName, val)

		default:
			fmt.Println("Request invalid metric type: ", r.URL.Path)
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		fmt.Println("Writing response: ", r.URL.Path)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

	}

}
