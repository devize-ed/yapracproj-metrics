package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/devize-ed/yapracproj-metrics.git/internal/compress"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
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

// middleware for compression coding
func MiddlewareGzip(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		ow := w // set original http.ResponseWriter

		// check if content type application/json or text/html
		contentType := r.Header.Get("Content-Type")
		isAccepted := strings.Contains(contentType, "application/json") ||
			strings.Contains(contentType, "text/html")
		if isAccepted {
			logger.Log.Debugf("Accepted Content Type for compression")
			// check if agent is assepting gzip and compress it
			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				logger.Log.Debugf("Agent support gzip encoding, compressing response...")
				cw := compress.NewCompressWriter(w)
				ow = cw
				defer cw.Close()
			}

			// check if received data compressed and decompress it
			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				logger.Log.Debugf("Received data is compressed, decompressing...")
				cr, err := compress.NewCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer cr.Close()
			}
		}

		h.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(logFn)
}
