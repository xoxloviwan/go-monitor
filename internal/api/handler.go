package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store ReaderWriter
}

type Reader interface {
	Get(metricType string, metricName string) (string, bool)
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
