package handler

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (h *Handler) NewRouter() http.Handler {
	// Initialize and configure the router, adding the route paths.
	r := chi.NewRouter()
	r.Use(MiddlewareLogging, MiddlewareGzip, middleware.StripSlashes)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateMetricHandler())
	r.Post("/update", h.UpdateMetricJSONHandler())
	r.Post("/value", h.GetMetricJSONHandler())
	r.Get("/value/{metricType}/{metricName}", h.GetMetricHandler())
	r.Get("/", h.ListMetricsHandler())

	return r
}
