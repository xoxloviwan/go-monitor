package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/xoxloviwan/go-monitor/internal/api"
	conf "github.com/xoxloviwan/go-monitor/internal/config_server"
)

const AddressDefault = "localhost:8080"

type Config struct {
	Address string `envDefault:"localhost:8080"`
}

func main() {
	cfg := conf.InitConfig()
	if len(flag.Args()) > 0 {
		fmt.Println("Too many arguments")
		os.Exit(1)
	}
	r := api.SetupRouter()
	r.Use(gin.Logger())
	err := r.Run(cfg.Address)
	if err != nil {
		panic(err)
	}
}
