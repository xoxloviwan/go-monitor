package main

import (
	"github.com/gin-gonic/gin"
	"github.com/xoxloviwan/go-monitor/internal/api"
)

func main() {
	handler := api.NewHandler()
	ginHandler := gin.WrapH(handler)
	r := gin.New()
	r.Use(ginHandler)
	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
}
