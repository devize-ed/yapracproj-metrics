package handler

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
)

// compressWriter wraps http.ResponseWriter to handle gzip compression.
type compressWriter struct {
	http.ResponseWriter
	zw       *gzip.Writer
	compress bool // flag if compression is needed
}

// newCompressWriter creates a new compressWriter that wraps the original http.ResponseWriter.
func newCompressWriter(w http.ResponseWriter, compress bool) *compressWriter {
	if !compress {
		return &compressWriter{ResponseWriter: w}
	}

	return &compressWriter{
		ResponseWriter: w,
		compress:       true,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.ResponseWriter.Header()
}

// WriteHeader writes the HTTP status code and sets the Content-Encoding header if compression is enabled.
func (c *compressWriter) WriteHeader(code int) {
	if c.compress {
		c.Header().Set("Content-Encoding", "gzip")
		c.Header().Del("Content-Length")
		if c.zw == nil {
			c.zw = gzip.NewWriter(c.ResponseWriter)
		}
	}
	c.ResponseWriter.WriteHeader(code)
	if c.compress && c.zw != nil {
		c.zw.Close()
	}
}
func (c *compressWriter) Write(b []byte) (int, error) {
	if !c.compress {
		return c.ResponseWriter.Write(b)
	}
	return c.zw.Write(b)
}

func (c *compressWriter) Close() error {
	if c.zw != nil {
		return c.zw.Close()
	}
	return nil
}

// compressReader wraps io.ReadCloser to handle gzip decompression.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader creates a new compressReader that decompresses gzip-encoded data.
func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// MiddlewareGzip is a middleware that handles gzip compression and decompression.
func MiddlewareGzip(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		ow := w // Set original http.ResponseWriter.
		logger.Log.Debug("req header", r.Header)
		// check if request is compressed, decompress it and remove Content-Encoding header.
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		logger.Log.Debugf("Content-Encoding: %s, sendsGzip: %t", contentEncoding, sendsGzip)
		if sendsGzip {
			logger.Log.Debugf("Received data is compressed,")
			cr, err := NewCompressReader(r.Body)
			if err != nil {
				logger.Log.Debugf("rror decompressing request: ", err)
				http.Error(w, "error decompressing request", http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()

			r.Header.Del("Content-Encoding")
			r.Header.Del("Content-Length")
		}

		// Check if agent is accepting gzip and compress it.
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		logger.Log.Debugf("Accept-Encoding: %s, gzip: %t", acceptEncoding, supportsGzip)
		if supportsGzip {
			cw := newCompressWriter(w, true)
			defer cw.Close()
			ow = cw
		}
		h.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(logFn)
}
