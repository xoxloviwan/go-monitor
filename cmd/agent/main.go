package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	metrs "github.com/xoxloviwan/go-monitor/internal/metrics"

	"github.com/caarlos0/env/v11"
)

const (
	AddressDefault        = "localhost:8080"
	PollIntervalDefault   = 2
	ReportIntervalDefault = 10
)

type Config struct {
	Address        string `envDefault:"localhost:8080"`
	ReportInterval int64  `envDefault:"10"`
	PollInterval   int64  `envDefault:"2"`
}

var (
	address              = flag.String("a", AddressDefault, "server adress")
	pollInterval         = flag.Int("p", PollIntervalDefault, "poll interval in seconds")
	reportInterval       = flag.Int("r", ReportIntervalDefault, "report interval in seconds")
	PollCount      int64 = 0
)

func send(adr *string, urls *[]string) (err error) {
	cl := &http.Client{}

	server := "http://" + *adr

	for _, url := range *urls {
		//fmt.Println(time.Now().Local().UTC(), "send", url)
		response, err := cl.Post(server+url, "text/plain", nil)
		if err != nil {
			return err
		}
		_, err = io.Copy(io.Discard, response.Body)
		defer response.Body.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	adr := address
	var cfg Config
	opts := env.Options{UseFieldNameByDefault: true}
	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		log.Fatalf("Error parsing env: %v", err)
	}
	flag.Parse()
	if len(flag.Args()) > 0 {
		log.Fatal("Too many arguments")
	}
	if cfg.Address != *address && cfg.Address != AddressDefault {
		adr = &cfg.Address
	}
	pollRate := int64(*pollInterval)
	if cfg.PollInterval != pollRate && cfg.PollInterval != PollIntervalDefault {
		pollRate = cfg.PollInterval
	}
	reportRate := int64(*reportInterval)
	if cfg.ReportInterval != reportRate && cfg.ReportInterval != ReportIntervalDefault {
		reportRate = cfg.ReportInterval
	}
	for {
		PollCount += 1
		metrics := metrs.GetMetrics(PollCount)
		if (PollCount*pollRate)%reportRate == 0 {
			urls := metrics.GetUrls()
			err := send(adr, &urls)
			if err != nil {
				fmt.Println(err)
			}
		}
		time.Sleep(time.Duration(pollRate) * time.Second)
	}
}
