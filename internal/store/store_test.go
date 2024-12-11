package store

import (
	"context"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	mtr "github.com/xoxloviwan/go-monitor/internal/metrics_types"
)

func setup(t *testing.T) *MemStorage {
	t.Helper()
	return NewMemStorage()
}

func TestMemStorage_AddGet(t *testing.T) {
	s := setup(t)
	// test Add api
	gaugeVal := 100.123
	gaugeValStr := strconv.FormatFloat(gaugeVal, 'f', -1, 64)
	err := s.Add("gauge", "test1", gaugeValStr)
	if err != nil {
		t.Errorf("MemStorage.Add() error = %v", err)
	}
	counterVal := int64(100)
	counterValStr := strconv.FormatInt(counterVal, 10)
	err = s.Add("counter", "test2", counterValStr)
	if err != nil {
		t.Errorf("MemStorage.Add() error = %v", err)
	}

	err = s.Add("undef", "undef", "undef")
	if err == nil {
		t.Error("MemStorage.Add(undef,undef,undef) found")
	}
	err = s.Add("gauge", "undef", "undef")
	if err == nil {
		t.Error("MemStorage.Add(gauge,undef,undef) must return error")
	}
	err = s.Add("counter", "undef", "undef")
	if err == nil {
		t.Error("MemStorage.Add(counter,undef,undef) must return error")
	}

	// check values inside MemStorage
	if s.Gauge["test1"] != gaugeVal {
		t.Errorf("MemStorage.Add() = %v, want %v", s.Gauge["test1"], gaugeVal)
	}
	if s.Counter["test2"] != counterVal {
		t.Errorf("MemStorage.Add() = %v, want %v", s.Counter["test2"], counterVal)
	}

	// test Get api
	gotGaugeVal, ok := s.Get("gauge", "test1")
	if !ok {
		t.Error("MemStorage.Get(gauge,test1) not found")
	}
	if gotGaugeVal != gaugeValStr {
		t.Errorf("MemStorage.Get() = %v, want %v", gotGaugeVal, gaugeVal)
	}
	gotCounterVal, ok := s.Get("counter", "test2")
	if !ok {
		t.Error("MemStorage.Get(counter,test2) not found")
	}
	if gotCounterVal != counterValStr {
		t.Errorf("MemStorage.Get() = %v, want %v", gotCounterVal, counterValStr)
	}

	_, ok = s.Get("undef", "undef")
	if ok {
		t.Error("MemStorage.Get(undef,undef) found")
	}
	_, ok = s.Get("counter", "undef")
	if ok {
		t.Error("MemStorage.Get(counter,undef) found")
	}
	_, ok = s.Get("gauge", "undef")
	if ok {
		t.Error("MemStorage.Get(gauge,undef) found")
	}
	storeString := s.String()
	if storeString == "" {
		t.Error("MemStorage.String() is empty")
	}
}

func TestMemStorage_AddMetricsGetMetrics(t *testing.T) {
	s := setup(t)
	metric1 := mtr.Metrics{
		ID:    "test1",
		MType: "gauge",
		Value: new(float64),
	}
	gaugeVal := 100.123
	*metric1.Value = gaugeVal
	metric2 := mtr.Metrics{
		ID:    "test2",
		MType: "counter",
		Delta: new(int64),
	}
	counterVal := int64(100)
	*metric2.Delta = counterVal

	metrics := &mtr.MetricsList{metric1, metric2}
	err := s.AddMetrics(context.Background(), metrics)
	if err != nil {
		t.Errorf("MemStorage.AddMetrics() error = %v", err)
	}
	wantMetrics := metrics

	if s.Gauge["test1"] != gaugeVal {
		t.Errorf("MemStorage.AddMetrics() = %v, want %v", s.Gauge["test1"], gaugeVal)
	}
	if s.Counter["test2"] != counterVal {
		t.Errorf("MemStorage.AddMetrics() = %v, want %v", s.Counter["test2"], counterVal)
	}

	metrics = &mtr.MetricsList{
		mtr.Metrics{
			ID:    "test1",
			MType: "gauge",
		},
		mtr.Metrics{
			ID:    "test2",
			MType: "counter",
		},
	}
	// test Get api
	gotMetrics, err := s.GetMetrics(context.Background(), *metrics)
	if err != nil {
		t.Errorf("MemStorage.GetMetrics() error = %v", err)
	}
	if diff := cmp.Diff(*wantMetrics, gotMetrics); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}

}
