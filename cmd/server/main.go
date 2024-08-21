package main

import (
	"log/slog"
	"os"

	"github.com/xoxloviwan/go-monitor/internal/api"
	conf "github.com/xoxloviwan/go-monitor/internal/config_server"
)

func main() {
	cfg := conf.InitConfig()
	err := api.RunServer(cfg)
	if err != nil {
		slog.Error("Server down", slog.Any("error", err))
		os.Exit(1)
	}
}
