package metrics

import (
	"math/rand"
	"runtime"

	"github.com/xoxloviwan/go-monitor/internal/api"
	"github.com/xoxloviwan/go-monitor/internal/store"
)

type MetricsPool store.MemStorage

func GetMetrics(PollCount int64) *MetricsPool {
	var MemStats runtime.MemStats
	runtime.ReadMemStats(&MemStats)
	return &MetricsPool{
		Gauge: map[string]float64{
			"Alloc":         float64(MemStats.Alloc),
			"BuckHashSys":   float64(MemStats.BuckHashSys),
			"Frees":         float64(MemStats.Frees),
			"GCCPUFraction": MemStats.GCCPUFraction,
			"GCSys":         float64(MemStats.GCSys),
			"HeapAlloc":     float64(MemStats.HeapAlloc),
			"HeapIdle":      float64(MemStats.HeapIdle),
			"HeapInuse":     float64(MemStats.HeapInuse),
			"HeapObjects":   float64(MemStats.HeapObjects),
			"HeapReleased":  float64(MemStats.HeapReleased),
			"HeapSys":       float64(MemStats.HeapSys),
			"LastGC":        float64(MemStats.LastGC),
			"Lookups":       float64(MemStats.Lookups),
			"MCacheInuse":   float64(MemStats.MCacheInuse),
			"MCacheSys":     float64(MemStats.MCacheSys),
			"MSpanInuse":    float64(MemStats.MSpanInuse),
			"MSpanSys":      float64(MemStats.MSpanSys),
			"Mallocs":       float64(MemStats.Mallocs),
			"NextGC":        float64(MemStats.NextGC),
			"NumForcedGC":   float64(MemStats.NumForcedGC),
			"NumGC":         float64(MemStats.NumGC),
			"OtherSys":      float64(MemStats.OtherSys),
			"PauseTotalNs":  float64(MemStats.PauseTotalNs),
			"StackInuse":    float64(MemStats.StackInuse),
			"StackSys":      float64(MemStats.StackSys),
			"Sys":           float64(MemStats.Sys),
			"TotalAlloc":    float64(MemStats.TotalAlloc),
			"RandomValue":   rand.Float64(),
		},
		Counter: map[string]int64{
			"PollCount": PollCount,
		},
	}
}

func (s *MetricsPool) MakeMessages() []api.Metrics {
	var msgs []api.Metrics
	for metricName, metricValue := range s.Gauge {
		msgs = append(msgs, api.Metrics{
			ID:    metricName,
			MType: store.GaugeName,
			Value: &metricValue,
		})
	}
	for metricName, metricValue := range s.Counter {
		msgs = append(msgs, api.Metrics{
			ID:    metricName,
			MType: store.CounterName,
			Delta: &metricValue,
		})
	}
	return msgs
}
