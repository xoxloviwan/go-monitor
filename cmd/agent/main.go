package main

import (
	"io"
	"log"
	"net/http"
	"time"

	conf "github.com/xoxloviwan/go-monitor/internal/config_agent"
	metrs "github.com/xoxloviwan/go-monitor/internal/metrics"
)

func send(adr *string, urls []string) (err error) {
	cl := &http.Client{}

	server := "http://" + *adr

	for _, url := range urls {
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
	pollTicker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	defer pollTicker.Stop()
	sendTicker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	defer sendTicker.Stop()
	// Нам не нужен глобальный счетчик, т.к. он используется только внутри функции main, поэтому его можно объявить внутри main.
	var pollCount int64
	// Получаем метрики сразу после инициализации. Таким образом метрики будут сразу доступны для отправки.
	metrics := metrs.GetMetrics(pollCount)
	for {
		// Здесь произойдет lock и select разлочится событием, которое произойдет первым.
		select {
		case <-pollTicker.C:
			pollCount += 1
			metrics = metrs.GetMetrics(pollCount)
		case <-sendTicker.C:
			urls := metrics.GetUrls()
			err := send(&cfg.Address, urls)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
