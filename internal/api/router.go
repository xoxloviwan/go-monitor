package api

import (
	"github.com/gin-gonic/gin"
	"github.com/xoxloviwan/go-monitor/internal/store"
)

func SetupRouter() *gin.Engine {
	store := store.NewMemStorage()
	handler := NewHandler(store)
	r := gin.New()
	r.Use(logger())
	r.Use(compressGzip())
	r.POST("/update/:metricType/:metricName/:metricValue", handler.update)
	r.POST("/update/", handler.updateJSON)
	r.GET("/value/:metricType/:metricName", handler.value)
	r.POST("/value/", handler.valueJSON)
	r.GET("/", handler.list)
	return r
}
