package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/gin-gonic/gin"
	"github.com/xoxloviwan/go-monitor/internal/api"
)

const AddressDefault = "localhost:8080"

type Config struct {
	Address string `envDefault:"localhost:8080"`
}

func main() {
	adr := flag.String("a", DA, "server adress")
	var cfg Config
	opts := env.Options{UseFieldNameByDefault: true}
	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	flag.Parse()
	if cfg.Address != *adr && cfg.Address != DA {
		adr = &cfg.Address
	}
	if len(flag.Args()) > 0 {
		fmt.Println("Too many arguments")
		os.Exit(1)
	}
	r := api.SetupRouter()
	r.Use(gin.Logger())
	err := r.Run(*adr)
	if err != nil {
		panic(err)
	}
}
