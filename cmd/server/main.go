package main

import (
	"github.com/gin-gonic/gin"
	"github.com/xoxloviwan/go-monitor/internal/api"
)

func main() {
	r := api.SetupRouter()
	r.Use(gin.Logger())
	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
}
