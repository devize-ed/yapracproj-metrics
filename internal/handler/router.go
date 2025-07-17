package handler

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (h *Handler) NewRouter() http.Handler {
	// init and configure the router, adding the route paths
	r := chi.NewRouter()
	r.Use(MiddlewareLogging, middleware.StripSlashes) //logging middleware
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateMetricHandler())
	r.Post("/update", h.UpdateMetricJsonHandler())
	r.Post("/value", h.GetMetricJsonHandler())
	r.Get("/value/{metricType}/{metricName}", h.GetMetricHandler())
	r.Get("/", h.ListMetricsHandler())

	return r
}
