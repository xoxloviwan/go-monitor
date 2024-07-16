package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	conf "github.com/xoxloviwan/go-monitor/internal/config"
	metrs "github.com/xoxloviwan/go-monitor/internal/metrics"
)

var (
	PollCount int64 = 0
)

func send(adr *string, urls *[]string) (err error) {
	cl := &http.Client{}

	server := "http://" + *adr

	for _, url := range *urls {
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
	cfg := conf.InitConfig()
	for {
		PollCount += 1
		metrics := metrs.GetMetrics(PollCount)
		if (PollCount*cfg.PollInterval)%cfg.ReportInterval == 0 {
			urls := metrics.GetUrls()
			err := send(&cfg.Address, &urls)
			if err != nil {
				fmt.Println(err)
			}
		}
		time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
	}
}
