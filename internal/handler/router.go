package handler

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (h *Handler) NewRouter() http.Handler {
	// init and configure the router, adding the route paths
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.AllowContentType("text/plain; charset=utf-8")) //logging and allow only text/plain content type from the agent
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateMetricHandler())
	r.Get("/value/{metricType}/{metricName}", h.GetMetricHandler())
	r.Get("/", h.ListMetricsHandler())

	return r
}
