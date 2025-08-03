package handler

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	mw "github.com/devize-ed/yapracproj-metrics.git/internal/handler/middleware"
)

func (h *Handler) NewRouter() http.Handler {
	// Initialize and configure the router, adding the route paths.
	r := chi.NewRouter()
	r.Use(mw.MiddlewareLogging, middleware.StripSlashes, mw.MiddlewareGzip)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateMetricHandler())
	r.Post("/update", h.UpdateMetricJSONHandler())
	r.Post("/updates", h.UpdateBatchHandler())
	r.Post("/value", h.GetMetricJSONHandler())
	r.Get("/value/{metricType}/{metricName}", h.GetMetricHandler())
	r.Get("/", h.ListMetricsHandler())
	r.Get("/ping", h.PingHandler())
	return r
}
