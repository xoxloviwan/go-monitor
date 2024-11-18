package config

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/xoxloviwan/go-monitor/internal/helpers"
)

const (
	addressDefault        = "localhost:8080"
	pollIntervalDefault   = 2
	reportIntervalDefault = 10
	rateLimitDefault      = 1
)

var (
	address        = flag.String("a", addressDefault, "server adress")
	pollInterval   = flag.Int("p", pollIntervalDefault, "poll interval in seconds")
	reportInterval = flag.Int("r", reportIntervalDefault, "report interval in seconds")
	key            = flag.String("k", "", "key for encrypting and decrypting data, e.g. 8c17b18522bf3f559864ac08f74c8ddb")
	cryptoKey      = flag.String("crypto-key", "", "path to file with public key for encrypting data")
	rateLimit      = flag.Int("l", rateLimitDefault, "number of outgoing requests at once")
	config         = flag.String("c", "", "path to config file")
)

// Config represents the configuration for the agent.
//
// It contains fields for the server address, report interval, poll interval, key, and rate limit.
type Config struct {
	// Address is the server address.
	Address string `envDefault:"localhost:8080"`
	// ReportInterval is the interval at which metrics are reported.
	ReportInterval int64 `envDefault:"10"`
	// PollInterval is the interval at which metrics are polled.
	PollInterval int64 `envDefault:"2"`
	// Key for signing of request body data by SHA256 algorithm
	Key string `envDefault:""`
	// Path to file with public key for ecrypting request body
	CryptoKey string `envDefault:""`
	// RateLimit is the number of outgoing requests at once.
	RateLimit int `envDefault:"1"`
}

// FileConfig represents the json configuration in file
type FileConfig struct {
	Config
	ReportIntervalT helpers.Duration `json:"report_interval"`
	PollIntervalT   helpers.Duration `json:"poll_interval"`
}

// ConfigFull represents the env configuration for the agent with path to config file
type ConfigFull struct {
	Config
	// Path to config file
	ConfigPath string `env:"CONFIG" envDefault:""`
}

// InitConfig initializes a new Config instance.
//
// The instance is initialized with the given environment variables and command-line flags.
func InitConfig() Config {
	cfgDefaults := Config{
		Address:        addressDefault,
		PollInterval:   pollIntervalDefault,
		ReportInterval: reportIntervalDefault,
		RateLimit:      rateLimitDefault,
		Key:            "",
		CryptoKey:      "",
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
		Address:        *address,
		PollInterval:   int64(*pollInterval),
		ReportInterval: int64(*reportInterval),
		RateLimit:      *rateLimit,
		Key:            *key,
		CryptoKey:      *cryptoKey,
	})
	redefineConf(&cfgDefaults, cfg.Config)
	log.Print(cfgDefaults)
	return cfgDefaults
}

func redefineConf(cfg *Config, leadCfg Config) {
	log.Println(cfg.Address)
	if cfg.Address != leadCfg.Address && leadCfg.Address != addressDefault {
		cfg.Address = leadCfg.Address
	}

	if cfg.PollInterval != leadCfg.PollInterval && leadCfg.PollInterval != pollIntervalDefault {
		cfg.PollInterval = leadCfg.PollInterval
	}

	if cfg.ReportInterval != leadCfg.ReportInterval && leadCfg.ReportInterval != reportIntervalDefault {
		cfg.ReportInterval = leadCfg.ReportInterval
	}
	if cfg.Key != leadCfg.Key && leadCfg.Key != "" {
		cfg.Key = leadCfg.Key
	}
	if cfg.CryptoKey != leadCfg.CryptoKey && leadCfg.CryptoKey != "" {
		cfg.CryptoKey = leadCfg.CryptoKey
	}
	if cfg.RateLimit != leadCfg.RateLimit && leadCfg.RateLimit != rateLimitDefault {
		cfg.RateLimit = leadCfg.RateLimit
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
	cfgFile.Config.ReportInterval = int64(cfgFile.ReportIntervalT.Seconds())
	cfgFile.Config.PollInterval = int64(cfgFile.PollIntervalT.Seconds())
	if !bytes.Contains(data, []byte("rate_limit")) {
		cfgFile.Config.RateLimit = rateLimitDefault
	}
	return cfgFile.Config
}
