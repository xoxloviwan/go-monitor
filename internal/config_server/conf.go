package config_server

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

const (
	AddressDefault = "localhost:8080"
)

var (
	address = flag.String("a", AddressDefault, "server adress")
)

type Config struct {
	Address string `envDefault:"localhost:8080"`
}

func InitConfig() Config {
	cfg := Config{}
	opts := env.Options{UseFieldNameByDefault: true}
	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		log.Fatalf("Error parsing env: %v", err)
	}
	flag.Parse()
	if len(flag.Args()) > 0 {
		log.Fatal("Too many arguments")
	}
	if cfg.Address != *address && cfg.Address == AddressDefault {
		cfg.Address = *address
	}
	return cfg
}
