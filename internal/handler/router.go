package handler

import (
	"net/http"

	"github.com/go-chi/chi"
)

func (h *Handler) NewRouter() http.Handler {
	// init and configure the router, adding the route paths
	r := chi.NewRouter()
	r.Use(MiddlewareLogging) //logging middleware
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateMetricHandler())
	r.Post("/update", h.UpdateMetricJsonHandler())
	r.Post("/value", h.GetMetricJsonHandler())
	r.Get("/value/{metricType}/{metricName}", h.GetMetricHandler())
	r.Get("/", h.ListMetricsHandler())

	return r
}
