package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
	mtrTypes "github.com/xoxloviwan/go-monitor/internal/metrics_types"
)

// Handler is an API handler.
//
// It contains a store that implements the ReaderWriter interface.
type Handler struct {
	store ReaderWriter
}

// Reader is an interface for reading metrics.
//
// It provides methods for getting metrics values and lists of metrics.
type Reader interface {
	Get(metricType string, metricName string) (string, bool)
	GetMetrics(ctx context.Context, metricList mtrTypes.MetricsList) (mtrTypes.MetricsList, error)
	String() string
}

// Writer is an interface for writing metrics.
//
// It provides methods for adding metrics values and lists of metrics.
type Writer interface {
	Add(metricType string, metricName string, metricValue string) error
	AddMetrics(ctx context.Context, m *mtrTypes.MetricsList) error
}

//go:generate mockgen -destination ./mock/mock_store.go github.com/xoxloviwan/go-monitor/internal/api ReaderWriter

// ReaderWriter is an interface that combines Reader and Writer.
//
// It provides methods for both reading and writing metrics.
type ReaderWriter interface {
	Reader
	Writer
}

// newHandler returns a new API handler.
//
// The handler is initialized with the given store, which must implement the ReaderWriter interface.
func newHandler(store ReaderWriter) *Handler {
	return &Handler{
		store: store,
	}
}

func (hdl *Handler) update(c *gin.Context) {
	metricType := c.Param("metricType")
	metricName := c.Param("metricName")
	metricValue := c.Param("metricValue")

	if metricName == "" {
		c.Status(http.StatusNotFound)
		return
	}

	err := hdl.store.Add(metricType, metricName, metricValue)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusBadRequest)
	} else {
		c.Status(http.StatusOK)
	}
}

func (hdl *Handler) updateJSON(c *gin.Context) {

	if c.Request.Header.Get("Content-Type") != "application/json" {
		c.Error(fmt.Errorf("invalid content type"))
		c.Status(http.StatusBadRequest)
		return
	}
	ctx := c.Request.Context()
	defer ctx.Done()

	var mtr mtrTypes.Metrics
	var mtrList mtrTypes.MetricsList
	var err error

	var buf bytes.Buffer
	tee := io.TeeReader(c.Request.Body, &buf)

	err = easyjson.UnmarshalFromReader(tee, &mtrList)
	if err != nil {
		err = easyjson.UnmarshalFromReader(&buf, &mtr)
		if err != nil {
			c.Error(err)
			c.Status(http.StatusBadRequest)
			return
		}
		mtrList = mtrTypes.MetricsList{mtr}
	}

	err = hdl.store.AddMetrics(ctx, &mtrList)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	err = ctx.Err()
	if err != nil {
		c.Error(err)
		c.Status(http.StatusBadRequest)
		return
	}
	var mtrListWithValues mtrTypes.MetricsList
	mtrListWithValues, err = hdl.store.GetMetrics(ctx, mtrList)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	if mtr.ID != "" {
		mtrUpd := mtrTypes.Metrics{
			ID:    mtrListWithValues[0].ID,
			MType: mtrListWithValues[0].MType,
			Value: mtrListWithValues[0].Value,
			Delta: mtrListWithValues[0].Delta,
		}
		_, err = easyjson.MarshalToWriter(&mtrUpd, c.Writer)
		if err != nil {
			c.Error(err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
		return
	}
	_, err = easyjson.MarshalToWriter(&mtrListWithValues, c.Writer)
	if err != nil {
		c.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func (hdl *Handler) value(c *gin.Context) {
	metricType := c.Param("metricType")
	metricName := c.Param("metricName")

	v, ok := hdl.store.Get(metricType, metricName)

	if !ok {
		c.Status(http.StatusNotFound)
	} else {
		c.String(http.StatusOK, v)
	}

}

func (hdl *Handler) valueJSON(c *gin.Context) {

	if c.Request.Header.Get("Content-Type") != "application/json" {
		c.Error(fmt.Errorf("invalid content type"))
		c.Status(http.StatusBadRequest)
		return
	}

	var mtr mtrTypes.Metrics
	var err error

	if err = easyjson.UnmarshalFromReader(c.Request.Body, &mtr); err != nil {
		c.Error(err)
		c.Status(http.StatusBadRequest)
		return
	}

	val, ok := hdl.store.Get(mtr.MType, mtr.ID)
	if !ok {
		c.Error(fmt.Errorf("metric %s in store not found", mtr.ID))
		c.Status(http.StatusNotFound)
		return
	}
	if mtr.MType == "counter" {
		mtr.Delta = new(int64)
		*mtr.Delta, _ = strconv.ParseInt(val, 10, 64)
	} else if mtr.MType == "gauge" {
		mtr.Value = new(float64)
		*mtr.Value, _ = strconv.ParseFloat(val, 64)
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	easyjson.MarshalToWriter(&mtr, c.Writer)
}

func (hdl *Handler) list(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	res := hdl.store.String()
	res = strings.ReplaceAll(res, "\n", "<br>")
	c.String(http.StatusOK, res)
}
