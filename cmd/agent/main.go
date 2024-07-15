package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/caarlos0/env/v11"
)

type metrics struct {
	Alloc         float64
	BuckHashSys   float64
	Frees         float64
	GCCPUFraction float64
	GCSys         float64
	HeapAlloc     float64
	HeapIdle      float64
	HeapInuse     float64
	HeapObjects   float64
	HeapReleased  float64
	HeapSys       float64
	LastGC        float64
	Lookups       float64
	MCacheInuse   float64
	MCacheSys     float64
	MSpanInuse    float64
	MSpanSys      float64
	Mallocs       float64
	NextGC        float64
	NumForcedGC   float64
	NumGC         float64
	OtherSys      float64
	PauseTotalNs  float64
	StackInuse    float64
	StackSys      float64
	Sys           float64
	TotalAlloc    float64
	RandomValue   float64
	PollCount     int64
}

const (
        AddressDefault = "localhost:8080"
        PollIntervalDefault = 2
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

func getMetrics(MemStats *runtime.MemStats, PollCount int64) metrics {
	return metrics{
		Alloc:         float64(MemStats.Alloc),
		BuckHashSys:   float64(MemStats.BuckHashSys),
		Frees:         float64(MemStats.Frees),
		GCCPUFraction: MemStats.GCCPUFraction,
		GCSys:         float64(MemStats.GCSys),
		HeapAlloc:     float64(MemStats.HeapAlloc),
		HeapIdle:      float64(MemStats.HeapIdle),
		HeapInuse:     float64(MemStats.HeapInuse),
		HeapObjects:   float64(MemStats.HeapObjects),
		HeapReleased:  float64(MemStats.HeapReleased),
		HeapSys:       float64(MemStats.HeapSys),
		LastGC:        float64(MemStats.LastGC),
		Lookups:       float64(MemStats.Lookups),
		MCacheInuse:   float64(MemStats.MCacheInuse),
		MCacheSys:     float64(MemStats.MCacheSys),
		MSpanInuse:    float64(MemStats.MSpanInuse),
		MSpanSys:      float64(MemStats.MSpanSys),
		Mallocs:       float64(MemStats.Mallocs),
		NextGC:        float64(MemStats.NextGC),
		NumForcedGC:   float64(MemStats.NumForcedGC),
		NumGC:         float64(MemStats.NumGC),
		OtherSys:      float64(MemStats.OtherSys),
		PauseTotalNs:  float64(MemStats.PauseTotalNs),
		StackInuse:    float64(MemStats.StackInuse),
		StackSys:      float64(MemStats.StackSys),
		Sys:           float64(MemStats.Sys),
		TotalAlloc:    float64(MemStats.TotalAlloc),
		RandomValue:   rand.Float64(),
		PollCount:     PollCount,
	}
}

func (m *metrics) getUrls() []string {
	var urls []string
	v := reflect.ValueOf(*m)
	for i := 0; i < v.NumField(); i++ {
		var url string
		if v.Type().Field(i).Type.String() == "int64" {
			url = "/update/counter/" + v.Type().Field(i).Name + "/" + fmt.Sprintf("%v", v.Field(i))
		} else {
			url = "/update/gauge/" + v.Type().Field(i).Name + "/" + fmt.Sprintf("%v", v.Field(i))
		}
		urls = append(urls, url)
	}
	return urls
}

func send(adr *string, urls *[]string) {
	cl := &http.Client{}

	server := "http://" + *adr

	for _, url := range *urls {
		//fmt.Println(time.Now().Local().UTC(), "send", url)
		response, err := cl.Post(server+url, "text/plain", nil)
		if err != nil {
			panic(err)
		}
		_, err = io.Copy(io.Discard, response.Body)
		defer response.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}

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
	var MemStats runtime.MemStats
	for {
		runtime.ReadMemStats(&MemStats)
		PollCount = PollCount + 1
		metrics := getMetrics(&MemStats, PollCount)

		if (PollCount*pollRate)%reportRate == 0 {
			urls := metrics.getUrls()
			send(adr, &urls)
		}
		time.Sleep(time.Duration(pollRate) * time.Second)
	}
}
