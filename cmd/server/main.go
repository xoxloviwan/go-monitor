package main

import (
	"log"

	"github.com/xoxloviwan/go-monitor/internal/api"
	conf "github.com/xoxloviwan/go-monitor/internal/config_server"
)

func main() {
	cfg := conf.InitConfig()
	err := api.RunServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
}
