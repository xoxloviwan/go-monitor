package main

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/mailru/easyjson"
	conf "github.com/xoxloviwan/go-monitor/internal/config_agent"
	metrs "github.com/xoxloviwan/go-monitor/internal/metrics"
	api "github.com/xoxloviwan/go-monitor/internal/metrics_types"
)

func send(workerId int, adr string, msgs api.MetricsList, key string) (err error) {
	cl := &http.Client{}

	url := "http://" + adr + "/updates/"

	var body []byte
	body, err = easyjson.Marshal(&msgs)
	if err != nil {
		return err
	}
	var sign string
	if key != "" {
		sign, err = getHash(body, key)
		if err != nil {
			return err
		}
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

	if key != "" {
		req.Header.Set("HashSHA256", sign)
	}

	// logReq := []any{
	// 	slog.String("url", url),
	// 	slog.String("body", string(body))}

	// for header, values := range req.Header {
	// 	logReq = append(logReq, slog.String(header, strings.Join(values, ",")))
	// }

	// slog.Info("REQ", logReq...)

	var response *http.Response
	retry := 0
	response, err = cl.Do(req)
	for err != nil && retry < 3 {
		if response != nil {
			response.Body.Close()
		}
		after := (retry+1)*2 - 1
		time.Sleep(time.Duration(after) * time.Second)
		log.Printf("worker #%d: %s Retry %d ...", workerId, err.Error(), retry+1)
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

func getHash(data []byte, strkey string) (string, error) {
	h := hmac.New(sha256.New, []byte(strkey))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}

	sign := h.Sum(nil)
	return hex.EncodeToString(sign), nil
}

// source - канал со снятыми метриками, содержит весь пакет
// n - на сколько запросов/работников можно разделить пакет метрик
func SplitBatch(source <-chan api.Metrics, n int) []<-chan api.Metrics {
	dests := make([]<-chan api.Metrics, 0) // Создать срез dests

	for i := 0; i < n; i++ { // Создать n выходных каналов
		ch := make(chan api.Metrics)
		dests = append(dests, ch)
		go func() { // Каждый выходной канал передается
			defer close(ch) // своей сопрограмме, которая состязается
			// с другими за доступ к source
			for val := range source {
				ch <- val
			}
		}()
	}
	return dests
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
			msgCh := metrics.MakeMessages()
			dests := SplitBatch(msgCh, cfg.RateLimit) // Fan Out

			var wg sync.WaitGroup // Использовать WaitGroup для ожидания, пока
			wg.Add(len(dests))    // не закроются выходные каналы
			for i, ch := range dests {
				go func(worker int, d <-chan api.Metrics) {
					defer wg.Done()
					subbatch := make([]api.Metrics, 0)
					for val := range d {
						subbatch = append(subbatch, val)
					}
					if len(subbatch) > 0 {
						log.Printf("worker #%d got %+v\n", worker, subbatch)
						err := send(worker, cfg.Address, subbatch, cfg.Key)
						if err != nil {
							log.Printf("worker #%d error: %s\n", worker, err.Error())
						}
					}
				}(i, ch)
			}
			wg.Wait()
			log.Println("Jobs done")
		}
	}
}
