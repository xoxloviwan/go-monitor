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
	r.POST("/update/:metricType/:metricName/:metricValue", handler.update)
	r.GET("/value/:metricType/:metricName", handler.value)
	return r
}
