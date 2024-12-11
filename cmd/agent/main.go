package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"log/slog"

	asc "github.com/xoxloviwan/go-monitor/internal/asymcrypto"
	"github.com/xoxloviwan/go-monitor/internal/clients"
	"github.com/xoxloviwan/go-monitor/internal/clients/base"
	conf "github.com/xoxloviwan/go-monitor/internal/config_agent"
	metrs "github.com/xoxloviwan/go-monitor/internal/metrics"
	api "github.com/xoxloviwan/go-monitor/internal/metrics_types"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

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
	localIP, _ := base.GetIP()
	if cfg.GRPC != "" {
		cfg.Address = cfg.GRPC
	}
	sender := clients.NewSender(cfg.GRPC != "", cfg.Address, cfg.Key, localIP.String(), publicKey)
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
						err := sender.Send(worker, subbatch)
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
