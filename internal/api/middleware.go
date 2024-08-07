package api

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type reqPars struct {
	URI      string
	method   string
	duration time.Duration
}

type respPars struct {
	code     int
	bodySize int
}

type logParams struct {
	reqPars
	respPars
}

func (l *logParams) String() string {
	return l.method + " - " + strconv.Itoa(l.code) + " - " + strconv.Itoa(l.bodySize) + " - " + l.duration.String() + " - " + l.URI
}

func logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Start timer
		start := time.Now()
		// Process request
		ctx.Next()

		pars := logParams{
			reqPars: reqPars{
				URI:      ctx.Request.URL.Path,
				method:   ctx.Request.Method,
				duration: time.Since(start),
			},
			respPars: respPars{
				code:     ctx.Writer.Status(),
				bodySize: ctx.Writer.Size(),
			},
		}

		slog.Info(pars.String())
	}
}

type compressWriter struct {
	gin.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w gin.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		zw:             gzip.NewWriter(w),
	}
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
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

func compressGzip() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		acceptEncoding := ctx.Request.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(ctx.Writer)
			cw.Header().Set("Content-Encoding", "gzip")
			ctx.Writer = cw
			defer cw.Close()
		}

		contentEncoding := ctx.Request.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(ctx.Request.Body)
			if err != nil {
				ctx.Writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer cr.Close()
			ctx.Request.Body = cr
		}

		ctx.Next()
	}
}
