package main

import (
	"github.com/gin-gonic/gin"
	"github.com/xoxloviwan/go-monitor/internal/api"
	"github.com/xoxloviwan/go-monitor/internal/store"
)

func main() {
	store := store.NewMemStorage()
	handler := api.NewHandler(store)
	ginHandler := gin.WrapH(handler)
	r := gin.New()
	r.Use(ginHandler)
	r.Use(gin.Logger())
	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
}
