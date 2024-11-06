package main

import (
	"fmt"
	"log/slog"

	"github.com/xoxloviwan/go-monitor/internal/api"
	conf "github.com/xoxloviwan/go-monitor/internal/config_server"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)
	cfg := conf.InitConfig()
	err := api.RunServer(cfg)
	if err != nil {
		api.LogFatal("Server down", slog.Any("error", err))
	}
}
