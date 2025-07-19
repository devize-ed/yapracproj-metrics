package handler

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
)

// compression f or the response
type compressWriter struct {
	http.ResponseWriter
	zw       *gzip.Writer
	compress bool // flag if compression is needed
}

func newCompressWriter(w http.ResponseWriter, compress bool) *compressWriter {
	if !compress {
		return &compressWriter{ResponseWriter: w}
	}

	zw := gzip.NewWriter(w)
	return &compressWriter{
		ResponseWriter: w,
		zw:             zw,
		compress:       true,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c *compressWriter) WriteHeader(code int) {
	// check Content‑Type and status code
	header := c.Header()
	if code < 300 && !c.compress {
		contentType := header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") ||
			strings.Contains(contentType, "text/html") {
			c.compress = true
		}
	}

	// compress if flag == true and set the header
	if c.compress && c.zw == nil {
		header.Set("Content-Encoding", "gzip")
		header.Del("Content-Length")
		c.zw = gzip.NewWriter(c.ResponseWriter)
	}
	c.ResponseWriter.WriteHeader(code)
}

func (c *compressWriter) Write(b []byte) (int, error) {
	if !c.compress {
		return c.ResponseWriter.Write(b)
	}

	if c.zw == nil {
		c.zw = gzip.NewWriter(c.ResponseWriter)
	}

	return c.zw.Write(b)
}

func (c *compressWriter) Close() error {
	if c.zw != nil {
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
		logger.Log.Debug("req header", r.Header)
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
			cw := newCompressWriter(w, true)
			w.Header().Set("Content-Encoding", "gzip")
			defer cw.Close()
			ow = cw
		}
		h.ServeHTTP(ow, r)
		fmt.Println(ow.Header())
	}

	return http.HandlerFunc(logFn)
}
