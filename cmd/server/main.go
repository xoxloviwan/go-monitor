package main

import (
	"log/slog"

	"github.com/xoxloviwan/go-monitor/internal/api"
	conf "github.com/xoxloviwan/go-monitor/internal/config_server"
)

func main() {
	cfg := conf.InitConfig()
	err := api.RunServer(cfg)
	if err != nil {
		api.LogFatal("Server down", slog.Any("error", err))
	}
}
