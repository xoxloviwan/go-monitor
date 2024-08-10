package api

import (
	"bytes"
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var Log *slog.Logger

var reqId = 0

func init() {
	Log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(Log)
}

func logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqId++

		// copy request body for logging
		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			bodyBytes = []byte(err.Error())
		}
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		Log.Info(
			"REQ",
			slog.Int("id", reqId),
			slog.String("method", ctx.Request.Method),
			slog.String("uri", ctx.Request.URL.Path),
			slog.Int64("body_size", ctx.Request.ContentLength),
			slog.String("ip", ctx.Request.RemoteAddr),
			slog.String("user_agent", ctx.Request.UserAgent()),
		)

		Log.Debug("REQ_BODY", slog.Int("id", reqId), slog.String("body", string(bodyBytes)))

		// Start timer
		start := time.Now()
		// Process request
		ctx.Next()

		status := ctx.Writer.Status()
		if status > 399 {
			Log.Error("RES",
				slog.Int("id", reqId),
				slog.Int("status", status),
				slog.Duration("duration", time.Since(start)),
				slog.String("err", ctx.Errors.String()),
			)
			return
		}
		Log.Info(
			"RES",
			slog.Int("id", reqId),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(start)),
			slog.Int("body_size", ctx.Writer.Size()),
		)
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
