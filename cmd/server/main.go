package main

import (
	"net/http"

	"log"

	"github.com/devize-ed/yapracproj-metrics.git/internal/handler"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const host = ":8080"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}

}

func run() error {
	storage := storage.NewMemStorage()

	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.AllowContentType("text/plain; charset=utf-8"))

	r.Post("/update/{metricType}/{metricName}/{metricValue}", handler.UpdateMetricHandler(storage))
	r.Get("/value/{metricType}/{metricName}", handler.GetMetricHandler(storage))
	r.Get("/", handler.ListAllHandler(storage))

	log.Println("Starting HTTP server on ", host)
	err := http.ListenAndServe(host, r)
	return err
}
