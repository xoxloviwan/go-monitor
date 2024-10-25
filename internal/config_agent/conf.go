package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

const (
	AddressDefault        = "localhost:8080"
	PollIntervalDefault   = 2
	ReportIntervalDefault = 10
	RateLimitDefault      = 1
)

var (
	address        = flag.String("a", AddressDefault, "server adress")
	pollInterval   = flag.Int("p", PollIntervalDefault, "poll interval in seconds")
	reportInterval = flag.Int("r", ReportIntervalDefault, "report interval in seconds")
	key            = flag.String("k", "", "key for encrypting and decrypting data, e.g. 8c17b18522bf3f559864ac08f74c8ddb")
	rateLimit      = flag.Int("l", RateLimitDefault, "number of outgoing requests at once")
)

type Config struct {
	Address        string `envDefault:"localhost:8080"`
	ReportInterval int64  `envDefault:"10"`
	PollInterval   int64  `envDefault:"2"`
	Key            string `envDefault:""`
	RateLimit      int    `envDefault:"1"`
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
	pollRate := int64(*pollInterval)
	if cfg.PollInterval != pollRate && cfg.PollInterval == PollIntervalDefault {
		cfg.PollInterval = pollRate
	}
	reportRate := int64(*reportInterval)
	if cfg.ReportInterval != reportRate && cfg.ReportInterval == ReportIntervalDefault {
		cfg.ReportInterval = reportRate
	}
	if cfg.Key != *key && cfg.Key == "" {
		cfg.Key = *key
	}
	if cfg.RateLimit != int(*rateLimit) && cfg.RateLimit == RateLimitDefault {
		cfg.RateLimit = int(*rateLimit)
	}
	return cfg

}
