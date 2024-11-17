package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/xoxloviwan/go-monitor/internal/helpers"
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
	key             = flag.String("k", "", "key for signing of response body data by SHA256 algorithm, e.g. 8c17b18522bf3f559864ac08f74c8ddb")
	cryptoKey       = flag.String("crypto-key", "", "path to file with private key for decrypting request body")
	config          = flag.String("c", "", "path to config file")
)

// Config represents the configuration for the server.
//
// It contains fields for the server address, store interval, file storage path, restore flag, database DSN, and key.
type Config struct {
	// Address is the server address.
	Address string `envDefault:"localhost:8080" json:"address"`
	// StoreInterval is the interval at which metrics are stored.
	StoreInterval int `envDefault:"300"`
	// FileStoragePath is the path to the file where metrics are stored.
	FileStoragePath string `envDefault:"" json:"file_storage_path"`
	// Restore is a flag indicating whether to restore data from file on startup.
	Restore bool `envDefault:"true" json:"restore"`
	// DatabaseDSN is the DSN for the database.
	DatabaseDSN string `envDefault:"" json:"database_dsn"`
	// Key  for signing of response body data by SHA256 algorithm
	Key string `envDefault:"" json:"key"`
	// Path to file with private key for decrypting request body
	CryptoKey string `envDefault:"" json:"crypto_key"`
}

type FileConfig struct {
	Config
	StoreIntervalT helpers.Duration `json:"store_interval"`
}

type ConfigFull struct {
	Config
	// Path to config file
	ConfigPath string `env:"CONFIG" envDefault:""`
}

// InitConfig initializes a new Config instance.
//
// Если указана переменная окружения, то используется она.
// Если нет переменной окружения, но есть аргумент командной строки (флаг), то используется он.
// Если нет ни переменной окружения, ни флага, то используется значение по умолчанию.
// Если есть файл конфигурации, то значения из файла конфигурации должны иметь меньший приоритет, чем флаги или переменные окружения.
// env > flag > file > default
func InitConfig() Config {
	cfgDefaults := Config{
		Address:         addressDefault,
		StoreInterval:   storeIntervalDefault,
		FileStoragePath: fileStoragePathDefault,
		Restore:         true,
		DatabaseDSN:     databaseDSNDefault,
		Key:             "",
		CryptoKey:       "",
	}
	cfg := ConfigFull{}
	opts := env.Options{UseFieldNameByDefault: true}
	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		log.Fatalf("Error parsing env: %v", err)
	}
	flag.Parse()
	if len(flag.Args()) > 0 {
		log.Fatal("Too many arguments")
	}
	if cfg.ConfigPath != *config && cfg.ConfigPath == "" {
		cfg.ConfigPath = *config
	}
	if cfg.ConfigPath != "" {
		cfgFile := configFromFile(cfg.ConfigPath)
		redefineConf(&cfgDefaults, cfgFile)
	}
	redefineConf(&cfgDefaults, Config{
		Address:         *address,
		StoreInterval:   *storeInterval,
		Restore:         *restore,
		FileStoragePath: *fileStoragePath,
		DatabaseDSN:     *databaseDSN,
		Key:             *key,
		CryptoKey:       *cryptoKey,
	})
	redefineConf(&cfgDefaults, cfg.Config)
	log.Print(cfgDefaults)
	return cfgDefaults
}

func redefineConf(cfg *Config, leadCfg Config) {
	log.Println(cfg.StoreInterval)
	if cfg.Address != leadCfg.Address && leadCfg.Address != addressDefault {
		cfg.Address = leadCfg.Address
	}

	if cfg.StoreInterval != leadCfg.StoreInterval && leadCfg.StoreInterval != storeIntervalDefault {
		cfg.StoreInterval = leadCfg.StoreInterval
	}

	if cfg.Restore != leadCfg.Restore && !leadCfg.Restore {
		cfg.Restore = false
	}

	if cfg.FileStoragePath != leadCfg.FileStoragePath && leadCfg.FileStoragePath != fileStoragePathDefault {
		cfg.FileStoragePath = leadCfg.FileStoragePath
	}
	if cfg.DatabaseDSN != leadCfg.DatabaseDSN && leadCfg.DatabaseDSN != databaseDSNDefault {
		cfg.DatabaseDSN = leadCfg.DatabaseDSN
	}

	if cfg.Key != leadCfg.Key && leadCfg.Key != "" {
		cfg.Key = leadCfg.Key
	}

	if cfg.CryptoKey != leadCfg.CryptoKey && leadCfg.CryptoKey != "" {
		cfg.CryptoKey = leadCfg.CryptoKey
	}
}

func configFromFile(path string) Config {
	var cfgFile FileConfig
	data, err := os.ReadFile(path)
	if err != nil {
		log.Println(err)
		return Config{}
	}
	err = json.Unmarshal(data, &cfgFile)
	if err != nil {
		log.Println(err)
		return Config{}
	}
	cfgFile.Config.StoreInterval = int(cfgFile.StoreIntervalT.Seconds())
	return cfgFile.Config
}
