package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/mailru/easyjson"
	"github.com/xoxloviwan/go-monitor/internal/api"
	conf "github.com/xoxloviwan/go-monitor/internal/config_agent"
	metrs "github.com/xoxloviwan/go-monitor/internal/metrics"
)

func send(adr *string, msgs []api.Metrics) (err error) {
	cl := &http.Client{}

	url := "http://" + *adr + "/update/"

	for _, msg := range msgs {
		body, err := easyjson.Marshal(&msg)
		if err != nil {
			return err
		}
		response, err := cl.Post(url, "application/json", bytes.NewBuffer(body))
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
			urls := metrics.MakeMessages()
			err := send(&cfg.Address, urls)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
