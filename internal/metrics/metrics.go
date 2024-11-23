package metrics

import (
	"fmt"
	"math/rand"
	"runtime"

	"log/slog"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"golang.org/x/sync/errgroup"

	api "github.com/xoxloviwan/go-monitor/internal/metrics_types"
	"github.com/xoxloviwan/go-monitor/internal/store"
)

// MetricsPool is a pool of metrics.
//
// It provides methods for getting metrics and making messages.
type MetricsPool store.MemStorage

type MetricsSync struct {
	vMem           *mem.VirtualMemoryStat
	cpuUtilization float64
	MemStats       runtime.MemStats
}

// GetMetrics returns a new MetricsPool instance.
//
// The instance is initialized with the given poll count.
func GetMetrics(PollCount int64) *MetricsPool {

	var tmp MetricsSync
	var eg errgroup.Group

	eg.Go(func() error {
		var MemStats runtime.MemStats
		runtime.ReadMemStats(&MemStats)
		tmp.MemStats = MemStats
		return nil
	})

	eg.Go(func() error {
		cpuUtilization, err := cpu.Percent(0, true) // вернет слайс с нагрузкой каждого ядра
		if err != nil {
			return fmt.Errorf("Getting cpu utilization failed: %w", err)
		}
		tmp.cpuUtilization = cpuUtilization[1]
		return nil
	})

	eg.Go(func() error {
		vMem, err := mem.VirtualMemory()
		if err != nil {
			return fmt.Errorf("Getting virtual memory failed: %w", err)
		}
		tmp.vMem = vMem
		return nil
	})

	if err := eg.Wait(); err != nil {
		slog.Error("Error getting metrics", "error", err)
	}
	return &MetricsPool{
		Gauge: map[string]float64{
			"Alloc":           float64(tmp.MemStats.Alloc),
			"BuckHashSys":     float64(tmp.MemStats.BuckHashSys),
			"Frees":           float64(tmp.MemStats.Frees),
			"GCCPUFraction":   tmp.MemStats.GCCPUFraction,
			"GCSys":           float64(tmp.MemStats.GCSys),
			"HeapAlloc":       float64(tmp.MemStats.HeapAlloc),
			"HeapIdle":        float64(tmp.MemStats.HeapIdle),
			"HeapInuse":       float64(tmp.MemStats.HeapInuse),
			"HeapObjects":     float64(tmp.MemStats.HeapObjects),
			"HeapReleased":    float64(tmp.MemStats.HeapReleased),
			"HeapSys":         float64(tmp.MemStats.HeapSys),
			"LastGC":          float64(tmp.MemStats.LastGC),
			"Lookups":         float64(tmp.MemStats.Lookups),
			"MCacheInuse":     float64(tmp.MemStats.MCacheInuse),
			"MCacheSys":       float64(tmp.MemStats.MCacheSys),
			"MSpanInuse":      float64(tmp.MemStats.MSpanInuse),
			"MSpanSys":        float64(tmp.MemStats.MSpanSys),
			"Mallocs":         float64(tmp.MemStats.Mallocs),
			"NextGC":          float64(tmp.MemStats.NextGC),
			"NumForcedGC":     float64(tmp.MemStats.NumForcedGC),
			"NumGC":           float64(tmp.MemStats.NumGC),
			"OtherSys":        float64(tmp.MemStats.OtherSys),
			"PauseTotalNs":    float64(tmp.MemStats.PauseTotalNs),
			"StackInuse":      float64(tmp.MemStats.StackInuse),
			"StackSys":        float64(tmp.MemStats.StackSys),
			"Sys":             float64(tmp.MemStats.Sys),
			"TotalAlloc":      float64(tmp.MemStats.TotalAlloc),
			"TotalMemory":     float64(tmp.vMem.Total),
			"FreeMemory":      float64(tmp.vMem.Free),
			"CPUutilization1": tmp.cpuUtilization,
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
				MType: store.GaugeName,
				Value: &metricValue,
			}
		}
		for metricName, metricValue := range s.Counter {
			ch <- api.Metrics{
				ID:    metricName,
				MType: store.CounterName,
				Delta: &metricValue,
			}
		}
	}()

	return ch
}
