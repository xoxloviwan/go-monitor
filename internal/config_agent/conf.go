package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
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
	rateLimit      = flag.Int("l", rateLimitDefault, "number of outgoing requests at once")
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
	// Key is the key used for encrypting and decrypting data.
	Key string `envDefault:""`
	// RateLimit is the number of outgoing requests at once.
	RateLimit int `envDefault:"1"`
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
	pollRate := int64(*pollInterval)
	if cfg.PollInterval != pollRate && cfg.PollInterval == pollIntervalDefault {
		cfg.PollInterval = pollRate
	}
	reportRate := int64(*reportInterval)
	if cfg.ReportInterval != reportRate && cfg.ReportInterval == reportIntervalDefault {
		cfg.ReportInterval = reportRate
	}
	if cfg.Key != *key && cfg.Key == "" {
		cfg.Key = *key
	}
	if cfg.RateLimit != int(*rateLimit) && cfg.RateLimit == rateLimitDefault {
		cfg.RateLimit = int(*rateLimit)
	}
	return cfg

}
