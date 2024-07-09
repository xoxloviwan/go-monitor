package main

import (
	"github.com/gin-gonic/gin"
	"github.com/xoxloviwan/go-monitor/internal/router"
)

func main() {
	r := router.SetupRouter()
	r.Use(gin.Logger())
	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
}
