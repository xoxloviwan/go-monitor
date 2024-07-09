package api

import (
	"github.com/gin-gonic/gin"
	"github.com/xoxloviwan/go-monitor/internal/store"
)

func SetupRouter() *gin.Engine {
	store := store.NewMemStorage()
	handler := NewHandler(store)
	ginHandler := gin.WrapH(handler)
	r := gin.New()
	r.Use(ginHandler)
	return r
}
