package metrics

import (
	"math/rand"
	"runtime"
	"sync"

	"log/slog"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	api "github.com/xoxloviwan/go-monitor/internal/metrics_types"
	"github.com/xoxloviwan/go-monitor/internal/store"
)

// MetricsPool is a pool of metrics.
//
// It provides methods for getting metrics and making messages.
type MetricsPool store.MemStorage

// GetMetrics returns a new MetricsPool instance.
//
// The instance is initialized with the given poll count.
func GetMetrics(PollCount int64) *MetricsPool {
	var wg sync.WaitGroup
	var cpuUtilization1 float64
	var MemStats runtime.MemStats
	var vMem *mem.VirtualMemoryStat
	wg.Add(3)

	go func() {
		runtime.ReadMemStats(&MemStats)
		wg.Done()
	}()

	go func() {
		cpuUtilization, err := cpu.Percent(0, true) // вернет слайс с нагрузкой каждого ядра
		if err != nil {
			slog.Error("Getting cpu utilization failed", "error", err)
			wg.Done()
			return
		}
		cpuUtilization1 = cpuUtilization[1]
		wg.Done()
	}()

	go func() {
		var err error
		vMem, err = mem.VirtualMemory()
		if err != nil {
			slog.Error("Getting virtual memory failed:", "error", err)
		}
		wg.Done()
	}()
	wg.Wait()
	return &MetricsPool{
		Gauge: map[string]float64{
			"Alloc":           float64(MemStats.Alloc),
			"BuckHashSys":     float64(MemStats.BuckHashSys),
			"Frees":           float64(MemStats.Frees),
			"GCCPUFraction":   MemStats.GCCPUFraction,
			"GCSys":           float64(MemStats.GCSys),
			"HeapAlloc":       float64(MemStats.HeapAlloc),
			"HeapIdle":        float64(MemStats.HeapIdle),
			"HeapInuse":       float64(MemStats.HeapInuse),
			"HeapObjects":     float64(MemStats.HeapObjects),
			"HeapReleased":    float64(MemStats.HeapReleased),
			"HeapSys":         float64(MemStats.HeapSys),
			"LastGC":          float64(MemStats.LastGC),
			"Lookups":         float64(MemStats.Lookups),
			"MCacheInuse":     float64(MemStats.MCacheInuse),
			"MCacheSys":       float64(MemStats.MCacheSys),
			"MSpanInuse":      float64(MemStats.MSpanInuse),
			"MSpanSys":        float64(MemStats.MSpanSys),
			"Mallocs":         float64(MemStats.Mallocs),
			"NextGC":          float64(MemStats.NextGC),
			"NumForcedGC":     float64(MemStats.NumForcedGC),
			"NumGC":           float64(MemStats.NumGC),
			"OtherSys":        float64(MemStats.OtherSys),
			"PauseTotalNs":    float64(MemStats.PauseTotalNs),
			"StackInuse":      float64(MemStats.StackInuse),
			"StackSys":        float64(MemStats.StackSys),
			"Sys":             float64(MemStats.Sys),
			"TotalAlloc":      float64(MemStats.TotalAlloc),
			"TotalMemory":     float64(vMem.Total),
			"FreeMemory":      float64(vMem.Free),
			"CPUutilization1": cpuUtilization1,
			"RandomValue":     rand.Float64(),
		},
		Counter: map[string]int64{
			"PollCount": PollCount,
		},
	}
}

// MakeMessages returns a channel of metrics messages.
//
// The channel is populated with messages from the MetricsPool instance.
func (s *MetricsPool) MakeMessages() chan api.Metrics {
	ch := make(chan api.Metrics)
	// через отдельную горутину генератор отправляет данные в канал
	go func() {
		// закрываем канал по завершению горутины — это отправитель
		defer close(ch)

		for metricName, metricValue := range s.Gauge {
			ch <- api.Metrics{
				ID:    metricName,
				MType: api.GaugeName,
				Value: &metricValue,
			}
		}
		for metricName, metricValue := range s.Counter {
			ch <- api.Metrics{
				ID:    metricName,
				MType: api.CounterName,
				Delta: &metricValue,
			}
		}
	}()

	return ch
}
