package api

import (
	"compress/gzip"
	"log/slog"
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

func compressGzip() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		acceptEncoding := ctx.Request.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(ctx.Writer)
			cw.Header().Set("Content-Encoding", "gzip")
			defer cw.Close()
			ctx.Writer = cw
		}
		ctx.Next()
	}
}
