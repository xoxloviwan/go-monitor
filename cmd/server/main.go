package main

import (
	"log"

	"github.com/xoxloviwan/go-monitor/internal/api"
	conf "github.com/xoxloviwan/go-monitor/internal/config_server"
)

func main() {
	cfg := conf.InitConfig()
	r := api.SetupRouter()
	err := r.Run(cfg.Address)
	if err != nil {
		log.Fatal(err)
	}
}
