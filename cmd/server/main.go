package main

import (
	"net/http"

	"log"

	"github.com/devize-ed/yapracproj-metrics.git/internal/handler"
	st "github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	parseFlags() // call the function to parse cl flags and get the host addr

	if err := run(); err != nil {
		log.Fatal(err)
	}

}

func run() error {
	ms := st.NewMemStorage() // init the memory storage for metrics

	// init and configure the router, adding the route paths
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.AllowContentType("text/plain; charset=utf-8")) //logging and allow only text/plain content type from the agent
	r.Post("/update/{metricType}/{metricName}/{metricValue}", handler.UpdateMetricHandler(ms))
	r.Get("/value/{metricType}/{metricName}", handler.GetMetricHandler(ms))
	r.Get("/", handler.ListAllHandler(ms))

	// init the http server
	srv := &http.Server{
		Addr:    host,
		Handler: r,
	}
	// loging the address and starting the server
	log.Println("Starting HTTP server on ", host)
	err := srv.ListenAndServe()
	return err
}
