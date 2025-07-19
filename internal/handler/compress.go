// package handler

// import (
// 	"compress/gzip"
// 	"io"
// 	"net/http"
// 	"strings"

// 	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
// )

// type compressWriter struct {
// 	w        http.ResponseWriter
// 	zw       *gzip.Writer
// 	compress bool // check if compression enaabled
// }

// func NewCompressWriter(w http.ResponseWriter) *compressWriter {
// 	return &compressWriter{
// 		w: w,
// 	}
// }

// func (c *compressWriter) Header() http.Header { return c.w.Header() }

// func (c *compressWriter) WriteHeader(statusCode int) {
// 	if statusCode < 300 {
// 		ct := c.w.Header().Get("Content-Type")
// 		if strings.HasPrefix(ct, "application/json") ||
// 			strings.HasPrefix(ct, "text/html") {

// 			c.compress = true
// 			c.w.Header().Set("Content-Encoding", "gzip")
// 			c.zw = gzip.NewWriter(c.w)
// 		}
// 	}
// 	c.w.WriteHeader(statusCode)
// }

// func (c *compressWriter) Write(p []byte) (int, error) {
// 	if !c.compress {
// 		return c.w.Write(p)
// 	}
// 	return c.zw.Write(p)
// }

// func (c *compressWriter) Close() error {
// 	if c.compress {
// 		return c.zw.Close()
// 	}
// 	return nil
// }

// // decompression, read compressed data
// type compressReader struct {
// 	r  io.ReadCloser
// 	zr *gzip.Reader
// }

// func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
// 	zr, err := gzip.NewReader(r)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &compressReader{
// 		r:  r,
// 		zr: zr,
// 	}, nil
// }

// func (c compressReader) Read(p []byte) (n int, err error) {
// 	return c.zr.Read(p)
// }

// func (c *compressReader) Close() error {
// 	if err := c.r.Close(); err != nil {
// 		return err
// 	}
// 	return c.zr.Close()
// }

// // middleware for compression coding
// func MiddlewareGzip(h http.Handler) http.Handler {
// 	logFn := func(w http.ResponseWriter, r *http.Request) {
// 		ow := w // set original http.ResponseWriter

// 		// check if received data compressed and decompress it
// 		contentEncoding := r.Header.Get("Content-Encoding")
// 		sendsGzip := strings.Contains(contentEncoding, "gzip")
// 		logger.Log.Debugf("Content-Encoding: %s, sendsGzip: %t", contentEncoding, sendsGzip)
// 		if sendsGzip {
// 			logger.Log.Debugf("Received data is compressed,")
// 			cr, err := NewCompressReader(r.Body)
// 			if err != nil {
// 				logger.Log.Debugf("rror decompressing request: ", err)
// 				http.Error(w, "error decompressing request", http.StatusInternalServerError)
// 				return
// 			}
// 			r.Body = cr
// 			defer cr.Close()

// 			r.Header.Del("Content-Encoding")
// 			r.Header.Del("Content-Length")
// 		}

// 		// check if agent is assepting gzip and compress it
// 		acceptEncoding := r.Header.Get("Accept-Encoding")
// 		supportsGzip := strings.Contains(acceptEncoding, "gzip")
// 		logger.Log.Debugf("Accept-Encoding: %s, gzip: %t", acceptEncoding, supportsGzip)
// 		if supportsGzip {
// 			logger.Log.Debugf("Agent support gzip encoding")
// 			cw := NewCompressWriter(w)
// 			ow = cw
// 			defer cw.Close()

// 			r.Header.Del("Content-Length")
// 		}

// 		h.ServeHTTP(ow, r)
// 	}
// 	return http.HandlerFunc(logFn)
// }

package handler

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
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
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(logFn)
}
