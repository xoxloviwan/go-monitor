package api

import (
	"log/slog"
	"strconv"
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
	return l.reqPars.method + " - " + strconv.Itoa(l.respPars.code) + " - " + strconv.Itoa(l.respPars.bodySize) + " - " + l.reqPars.duration.String() + " - " + l.reqPars.URI
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
