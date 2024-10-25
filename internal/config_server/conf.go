package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

const (
	addressDefault         = "localhost:8080"
	storeIntervalDefault   = 300
	fileStoragePathDefault = ""
	databaseDSNDefault     = ""
)

var (
	address         = flag.String("a", addressDefault, "server adress")
	storeInterval   = flag.Int("i", storeIntervalDefault, "store interval in seconds")
	fileStoragePath = flag.String("f", fileStoragePathDefault, "path to file with metrics")
	restore         = flag.Bool("r", true, "if need to restore data on start")
	databaseDSN     = flag.String("d", databaseDSNDefault, "database DSN, e.g. postgresql://postgres:12345@localhost:5432/postgres?sslmode=disable")
	key             = flag.String("k", "", "key for encrypting and decrypting data, e.g. 8c17b18522bf3f559864ac08f74c8ddb")
)

// Config represents the configuration for the server.
//
// It contains fields for the server address, store interval, file storage path, restore flag, database DSN, and key.
type Config struct {
	// Address is the server address.
	Address string `envDefault:"localhost:8080"`
	// StoreInterval is the interval at which metrics are stored.
	StoreInterval int `envDefault:"300"`
	// FileStoragePath is the path to the file where metrics are stored.
	FileStoragePath string `envDefault:""`
	// Restore is a flag indicating whether to restore data from file on startup.
	Restore bool `envDefault:"true"`
	// DatabaseDSN is the DSN for the database.
	DatabaseDSN string `envDefault:""`
	// Key is the key used for encryption and decryption.
	Key string `envDefault:""`
}

// InitConfig initializes a new Config instance.
//
// The instance is initialized with the given environment variables and command-line flags.
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
	if cfg.DatabaseDSN != *databaseDSN && cfg.DatabaseDSN == databaseDSNDefault {
		cfg.DatabaseDSN = *databaseDSN
	}

	if cfg.Key != *key && cfg.Key == "" {
		cfg.Key = *key
	}
	return cfg
}
