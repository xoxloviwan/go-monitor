package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

const (
	addressDefault         = "localhost:8080"
	storeIntervalDefault   = 300
	fileStoragePathDefault = "metrics.json"
	DatabaseDSNDefault     = "postgresql://postgres:12345@localhost:5432/postgres?sslmode=disable"
)

var (
	address         = flag.String("a", addressDefault, "server adress")
	storeInterval   = flag.Int("i", storeIntervalDefault, "store interval in seconds")
	fileStoragePath = flag.String("f", fileStoragePathDefault, "path to file with metrics")
	restore         = flag.Bool("r", true, "if need to restore data on start")
	databaseDSN     = flag.String("d", DatabaseDSNDefault, "database DSN")
)

type Config struct {
	Address         string `envDefault:"localhost:8080"`
	StoreInterval   int    `envDefault:"300"`
	FileStoragePath string `envDefault:"metrics.json"`
	Restore         bool   `envDefault:"true"`
	DatabaseDSN     string `envDefault:"postgresql://postgres:12345@localhost:5432/postgres?sslmode=disable"`
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
	if cfg.Address != *address && cfg.Address == addressDefault {
		cfg.Address = *address
	}

	if cfg.StoreInterval != *storeInterval && cfg.StoreInterval == storeIntervalDefault {
		cfg.StoreInterval = *storeInterval
	}

	if cfg.Restore != *restore {
		cfg.Restore = false
	}

	if cfg.FileStoragePath != *fileStoragePath && cfg.FileStoragePath == fileStoragePathDefault {
		cfg.FileStoragePath = *fileStoragePath
	}
	if cfg.DatabaseDSN != *databaseDSN && cfg.DatabaseDSN == DatabaseDSNDefault {
		cfg.DatabaseDSN = *databaseDSN
	}
	return cfg
}
