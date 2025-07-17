package handler

import (
	"net/http"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"go.uber.org/zap"
)

type (
	// Hold the status code and size of the response
	responseData struct {
		status int
		size   int
	}

	// custom ResponseWriter with capture of response data
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Redefine the Write and WriteHeader methods to capture response data
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// middleware for logging HTTP requests (URI, method, processing time, response status, size)
func MiddlewareLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Info("ReceivedHTTP request:",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		logger.Log.Infow("ReceivedHTTP request:",
			"uri", r.URL.Path,
			"method", r.Method,
			"body", r.Body,
		)

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

		logger.Log.Infow("Request info:",
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", duration,
		)

		logger.Log.Infow("Response info:",
			"status", responseData.status,
			"size", responseData.size,
		)

	}
	return http.HandlerFunc(logFn)
}
