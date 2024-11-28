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
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"log/slog"

	"github.com/mailru/easyjson"
	asc "github.com/xoxloviwan/go-monitor/internal/asymcrypto"
	conf "github.com/xoxloviwan/go-monitor/internal/config_agent"
	metrs "github.com/xoxloviwan/go-monitor/internal/metrics"
	api "github.com/xoxloviwan/go-monitor/internal/metrics_types"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// send
//
// workerID - идентификатор потока
// adr - адрес сервера
// msgs - список метрик
// key - ключ подписи
// publicKey - RSA публичный ключ для шифрования сообщения
func send(workerID int, adr string, msgs api.MetricsList, key string, publicKey *asc.PublicKey) (err error) {
	cl := &http.Client{}

	url := "http://" + adr + "/updates/"

	var body []byte
	body, err = easyjson.Marshal(&msgs)
	if err != nil {
		return err
	}
	slog.Info("got", "worker", workerID, "body", body)
	var sessionKey []byte
	if publicKey != nil {
		var err error
		sessionKey, body, err = asc.Encrypt(publicKey, body)
		if err != nil {
			return err
		}
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
		slog.Info("got", "worker", workerID, "body", body)
		return err
	}
	var req *http.Request
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(gzbody))
	if err != nil {
		return err
	}
	// net.IPAddr

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	// req.Header.Set("X-Real-IP")
	if sessionKey != nil {
		req.Header.Set("X-Key", hex.EncodeToString(sessionKey))
	}

	if key != "" {
		req.Header.Set("HashSHA256", sign)
	}

	var response *http.Response
	retry := 0
	response, err = cl.Do(req)
	for err != nil && retry < 3 {
		if response != nil {
			response.Body.Close()
		}
		after := (retry+1)*2 - 1
		time.Sleep(time.Duration(after) * time.Second)
		slog.Warn("Retry attempt", "worker", workerID, "error", err, "retry", retry+1)
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
// poolSize - на сколько запросов/работников можно разделить пакет метрик
func SplitBatch(source <-chan api.Metrics, poolSize int) []<-chan api.Metrics {
	dests := make([]<-chan api.Metrics, poolSize) // Создать массив dests

	for i := 0; i < poolSize; i++ { // Создать n выходных каналов
		ch := make(chan api.Metrics)
		dests[i] = ch
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
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)
	cfg := conf.InitConfig()
	var publicKey *asc.PublicKey
	if cfg.CryptoKey != "" {
		var err error
		publicKey, err = asc.GetPublicKey(cfg.CryptoKey)
		if err != nil {
			slog.Error("Error getting public key", "error", err)
			publicKey = nil
		}
	}
	pollTicker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	defer pollTicker.Stop()
	sendTicker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	defer sendTicker.Stop()
	// Нам не нужен глобальный счетчик, т.к. он используется только внутри функции main, поэтому его можно объявить внутри main.
	var pollCount int64
	// Получаем метрики сразу после инициализации. Таким образом метрики будут сразу доступны для отправки.
	metrics := metrs.GetMetrics(pollCount)
	var wg sync.WaitGroup // Используем WaitGroup для ожидания, пока не закроются выходные каналы
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for {
		// Здесь произойдет lock и select разлочится событием, которое произойдет первым.
		select {
		case <-pollTicker.C:
			pollCount += 1
			metrics = metrs.GetMetrics(pollCount)
		case <-sendTicker.C:
			msgCh := metrics.MakeMessages()
			dests := SplitBatch(msgCh, cfg.RateLimit) // Fan Out

			wg.Add(len(dests))
			for i, ch := range dests {
				go func(worker int, d <-chan api.Metrics) {
					defer wg.Done()
					subbatch := make([]api.Metrics, 0)
					for val := range d {
						subbatch = append(subbatch, val)
					}
					if len(subbatch) > 0 {
						slog.Info("Worker got task", "worker", worker, "subbatch", subbatch)
						err := send(worker, cfg.Address, subbatch, cfg.Key, publicKey)
						if err != nil {
							slog.Error("Send error", "worker", worker, "error", err)
						}
					}
				}(i, ch)
			}
			wg.Wait()
			slog.Info("Jobs done")
		case <-quit:
			slog.Info("Shutdown signal received...")
			wg.Wait()
			return
		}
	}
}

// func getIP() (net.IP, error) {
// 	conn, err := net.Dial("udp", "8.8.8.8:80")
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	defer conn.Close()

// 	localAddr := conn.LocalAddr().(*net.UDPAddr)

// 	return localAddr.IP, nil
// }

func getIP() (net.IP, error) {

	ifs, err := net.Interfaces()
	if err != nil {
		return []byte{}, err
	}
	var ip net.IP
	for _, itf := range ifs {
		addrs, err := itf.Addrs()
		if err != nil {
			return []byte{}, err
		}
		for _, address := range addrs {
			// check the address type and if it is not a loopback the display it
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ip4 := ipnet.IP.To4(); ip4 != nil {
					ip = ip4
					fmt.Println(ip, "\t\t\t", itf.Name)
				}
			}
		}
	}
	return ip, nil
	// return []byte{}, errors.New("are you connected to the network?")
}
