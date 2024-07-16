package main

import (
	"github.com/gin-gonic/gin"
	"github.com/xoxloviwan/go-monitor/internal/api"
	conf "github.com/xoxloviwan/go-monitor/internal/config_server"
)

func main() {
	cfg := conf.InitConfig()
	r := api.SetupRouter()
	r.Use(gin.Logger())
	err := r.Run(cfg.Address)
	if err != nil {
		panic(err)
	}
}
