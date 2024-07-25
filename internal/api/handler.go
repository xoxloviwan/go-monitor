package api

import (
	"net/http"
	"strconv"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
)

type Handler struct {
	store ReaderWriter
}

type Reader interface {
	Get(metricType string, metricName string) (string, bool)
	String() string
}

type Writer interface {
	Add(metricType string, metricName string, metricValue string) error
}

type ReaderWriter interface {
	Reader
	Writer
}

func NewHandler(store ReaderWriter) *Handler {
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
		c.Status(http.StatusBadRequest)
	} else {
		c.Status(http.StatusOK)
	}
}

func (hdl *Handler) updateJSON(c *gin.Context) {

	if c.Request.Header.Get("Content-Type") != "application/json" {
		c.Status(http.StatusBadRequest)
		return
	}

	var mtr Metrics
	var metricValue string

	if err := easyjson.UnmarshalFromReader(c.Request.Body, &mtr); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if mtr.Delta != nil {
		metricValue = strconv.FormatInt(*mtr.Delta, 10)
	}
	if mtr.Value != nil {
		metricValue = strconv.FormatFloat(*mtr.Value, 'f', -1, 64)
	}
	if mtr.Value == nil && mtr.Delta == nil {
		c.Status(http.StatusBadRequest)
		return
	}

	err := hdl.store.Add(mtr.MType, mtr.ID, metricValue)
	if err != nil {
		c.Status(http.StatusBadRequest)
	}
	val, _ := hdl.store.Get(mtr.MType, mtr.ID)

	mtrUpd := Metrics{
		ID:    mtr.ID,
		MType: mtr.MType,
	}

	if mtr.MType == "counter" {
		mtrUpd.Delta = new(int64)
		*mtrUpd.Delta, _ = strconv.ParseInt(val, 10, 64)
	} else if mtr.MType == "gauge" {
		mtrUpd.Value = new(float64)
		*mtrUpd.Value, _ = strconv.ParseFloat(val, 64)
	}
	c.Writer.Header().Add("Content-Type", "application/json")
	easyjson.MarshalToWriter(&mtrUpd, c.Writer)
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
		c.Status(http.StatusBadRequest)
		return
	}

	var mtr Metrics

	if err := easyjson.UnmarshalFromReader(c.Request.Body, &mtr); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	val, ok := hdl.store.Get(mtr.MType, mtr.ID)
	if !ok {
		c.Status(http.StatusNotFound)
	}
	if mtr.MType == "counter" {
		mtr.Delta = new(int64)
		*mtr.Delta, _ = strconv.ParseInt(val, 10, 64)
	} else if mtr.MType == "gauge" {
		mtr.Value = new(float64)
		*mtr.Value, _ = strconv.ParseFloat(val, 64)
	}
	c.Writer.Header().Add("Content-Type", "application/json")
	easyjson.MarshalToWriter(&mtr, c.Writer)
}

func (hdl *Handler) list(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	res := hdl.store.String()
	res = strings.ReplaceAll(res, "\n", "<br>")
	c.String(http.StatusOK, res)
}
