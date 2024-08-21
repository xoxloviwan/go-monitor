package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/mailru/easyjson"
	conf "github.com/xoxloviwan/go-monitor/internal/config_agent"
	metrs "github.com/xoxloviwan/go-monitor/internal/metrics"
	api "github.com/xoxloviwan/go-monitor/internal/metrics_types"
)

func send(adr string, msgs api.MetricsList) (err error) {
	cl := &http.Client{}

	url := "http://" + adr + "/updates/"

	var body []byte
	body, err = easyjson.Marshal(&msgs)
	if err != nil {
		return err
	}
	var gzbody []byte
	gzbody, err = compressGzip(body)
	if err != nil {
		return err
	}
	var req *http.Request
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(gzbody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	var response *http.Response
	retry := 0
	response, err = cl.Do(req)
	for err != nil && retry < 3 {
		if response != nil {
			response.Body.Close()
		}
		after := (retry+1)*2 - 1
		time.Sleep(time.Duration(after) * time.Second)
		log.Printf("%s Retry %d ...", err.Error(), retry+1)
		response, err = cl.Do(req)
		retry++
	}
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			closeErr = fmt.Errorf("could not close response body: %w", closeErr)
			err = errors.Join(err, closeErr)
		}
	}()

	_, err = io.Copy(io.Discard, response.Body)
	if err != nil {
		return err
	}
	return nil
}

// Compress сжимает слайс байт.
func compressGzip(data []byte) ([]byte, error) {
	var b bytes.Buffer
	// создаём переменную w — в неё будут записываться входящие данные,
	// которые будут сжиматься и сохраняться в bytes.Buffer
	w := gzip.NewWriter(&b)
	// запись данных
	_, err := w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}
	// обязательно нужно вызвать метод Close() — в противном случае часть данных
	// может не записаться в буфер b; если нужно выгрузить все упакованные данные
	// в какой-то момент сжатия, используйте метод Flush()
	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}
	// переменная b содержит сжатые данные
	return b.Bytes(), nil
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
			msgs := metrics.MakeMessages()
			err := send(cfg.Address, msgs)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
