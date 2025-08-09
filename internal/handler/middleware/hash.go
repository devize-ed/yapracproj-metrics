package handler

import (
	"bytes"
	"io"
	"net/http"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/devize-ed/yapracproj-metrics.git/internal/sign"
)

type hashResponseWriter struct {
	header http.Header
	buf    bytes.Buffer
	code   int
}

func newHashResponseWriter() *hashResponseWriter {
	return &hashResponseWriter{
		header: make(http.Header),
	}
}

func (h *hashResponseWriter) Header() http.Header {
	return h.header
}

func (h *hashResponseWriter) Write(b []byte) (int, error) {
	return h.buf.Write(b)
}

func (h *hashResponseWriter) WriteHeader(status int) {
	if h.code == 0 {
		h.code = status
	}
}

// HashMiddleware is a middleware that verifies the hash of the request body.
func HashMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Log.Debugf("Hash middleware with key %s", key)
			// If the key is empty, skip the hash verification.
			if key == "" {
				logger.Log.Debugf("key is empty")
				next.ServeHTTP(w, r)
				return
			}
			// Get the hash from the header.
			hash := r.Header.Get(sign.HashHeader)
			if hash == "" {
				// If the hash header is empty, skip the hash verification.
				logger.Log.Debugf("hash header is empty, %s", r.Header)
				http.Error(w, "missing hash header", http.StatusBadRequest)
				return
			}

			// Read the body of the request.
			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Debugf("Error reading request body: %v", err)
				http.Error(w, "Error reading request body", http.StatusBadRequest)
				return
			}
			// Close the original body to release the underlying reader before replacing it.
			_ = r.Body.Close()
			// Restore the body so other handlers can read it.
			r.Body = io.NopCloser(bytes.NewReader(body))

			// Verify the hash of the request body.
			ok, err := sign.Verify(body, key, hash)
			if err != nil || !ok {
				http.Error(w, "Hash verification failed", http.StatusBadRequest)
				return
			}

			logger.Log.Debugf("Hash verification passed")
			// Hash the response body.
			hw := newHashResponseWriter()
			next.ServeHTTP(hw, r)
			hw.WriteTo(w, key)
			logger.Log.Debugf("Hash response written")
		})
	}
}

func (h *hashResponseWriter) WriteTo(w http.ResponseWriter, key string) {
	// Copy the headers from the response writer.
	for k, vv := range h.header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	// Remove the Content-Length header.
	w.Header().Del("Content-Length")
	// Set the hash of the response body.
	if key != "" {
		w.Header().Set(sign.HashHeader, sign.Hash(h.buf.Bytes(), key))
	}
	// Set the status code.
	if h.code == 0 {
		h.code = http.StatusOK
	}
	// Write the response body.
	w.WriteHeader(h.code)
	_, _ = w.Write(h.buf.Bytes())
}
