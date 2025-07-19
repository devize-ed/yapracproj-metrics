package handler

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
)

type compressWriter struct {
	w        http.ResponseWriter
	zw       *gzip.Writer
	compress bool // check if compression enaabled
}

func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w: w,
	}
}

func (c *compressWriter) Header() http.Header { return c.w.Header() }

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		ct := c.w.Header().Get("Content-Type")
		if strings.HasPrefix(ct, "application/json") ||
			strings.HasPrefix(ct, "text/html") {

			c.compress = true
			c.w.Header().Set("Content-Encoding", "gzip")
			c.zw = gzip.NewWriter(c.w)
		}
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if !c.compress {
		return c.w.Write(p)
	}
	return c.zw.Write(p)
}

func (c *compressWriter) Close() error {
	if c.compress {
		return c.zw.Close()
	}
	return nil
}

// decompression, read compressed data
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

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

// middleware for compression coding
func MiddlewareGzip(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		ow := w // set original http.ResponseWriter

		// check if received data compressed and decompress it
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

		// check if agent is assepting gzip and compress it
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		logger.Log.Debugf("Accept-Encoding: %s, gzip: %t", acceptEncoding, supportsGzip)
		if supportsGzip {
			logger.Log.Debugf("Agent support gzip encoding")
			cw := NewCompressWriter(w)
			ow = cw
			defer cw.Close()

			r.Header.Del("Content-Length")
		}

		h.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(logFn)
}
