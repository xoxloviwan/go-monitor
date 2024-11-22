package metrics_test

import (
	"testing"

	m "github.com/xoxloviwan/go-monitor/internal/metrics"
)

var gauges = []string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
	"TotalMemory",
	"FreeMemory",
	"CPUutilization1",
	"RandomValue",
}

func TestGetMetrics(t *testing.T) {
	metrics := m.GetMetrics(1)

	for _, gauge := range gauges {
		if _, ok := metrics.Gauge[gauge]; !ok {
			t.Errorf("%s not exist", gauge)
		}
	}
	if metrics.Counter["PollCount"] != 1 {
		t.Errorf("PollCount wrong")
	}
}

func TestMakeMessage(t *testing.T) {
	mp := m.GetMetrics(1)
	ch := mp.MakeMessages()
	for msg := range ch {
		if msg.ID == "Alloc" {
			if msg.MType != "gauge" {
				t.Errorf("wrong MType %s", msg.MType)
			}
		}
	}
}
