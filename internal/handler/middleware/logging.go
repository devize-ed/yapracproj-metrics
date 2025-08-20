package handler

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	// Hold the status code and size of the response.
	responseData struct {
		status int
		size   int
	}

	// Custom ResponseWriter with capture of response data.
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Redefine the Write and WriteHeader methods to capture response data.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// MiddlewareLogging is a middleware for logging HTTP requests (URI, method, processing time, response status, size)
func MiddlewareLogging(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			responseData := &responseData{
				status: 0,
				size:   0,
			}

			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}
			h.ServeHTTP(&lw, r)

			duration := time.Since(start)

			logger.Infow("Request info:",
				"uri", r.RequestURI,
				"method", r.Method,
				"duration", duration,
			)

			logger.Infow("Response info:",
				"status", responseData.status,
				"size", responseData.size,
			)

		}
		return http.HandlerFunc(logFn)
	}
}
