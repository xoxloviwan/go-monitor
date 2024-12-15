package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/mailru/easyjson"
	mtr "github.com/xoxloviwan/go-monitor/internal/metrics_types"
)

//go:generate easyjson -output_filename store_easyjson_generated.go -all store.go

// CounterName is a constant representing the counter metric type.
const CounterName = mtr.CounterName

// GaugeName is a constant representing the gauge metric type.
const GaugeName = mtr.GaugeName

// Gauge is a map of gauge metrics.
//
// The keys are the metric names and the values are the metric values.
type Gauge map[string]float64

// Counter is a map of counter metrics.
//
// The keys are the metric names and the values are the metric values.
type Counter map[string]int64

// easyjson:json

// MemStorage is an in-memory storage implementation.
//
// It provides methods for adding metrics, getting metrics, restoring data from a file, and saving data to a file.
type MemStorage struct {
	Gauge   `json:"gauge"`
	Counter `json:"counter"`
}

// NewMemStorage returns a new MemStorage instance.
//
// The instance is initialized with empty Gauge and Counter maps.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

// Add adds a metric to the MemStorage instance.
//
// The metric is added with the given type, name, and value.
func (s *MemStorage) Add(metricType string, metricName string, metricValue string) (err error) {
	switch metricType {
	case CounterName:
		res64, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return err
		}
		s.Counter[metricName] += res64

	case GaugeName:
		res64, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return err
		}
		s.Gauge[metricName] = res64
	default:
		return errors.New("unknown metric type")
	}
	return err
}

// AddMetrics adds multiple metrics to the MemStorage instance.
//
// The metrics are added with the given context and metrics list.
func (s *MemStorage) AddMetrics(ctx context.Context, m *mtr.MetricsList) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	for _, v := range *m {
		if v.MType == GaugeName {
			s.Gauge[v.ID] = *v.Value
		}
		if v.MType == CounterName {
			s.Counter[v.ID] = *v.Delta + s.Counter[v.ID]
		}
	}
	return nil
}

// GetMetrics gets metrics from the MemStorage instance.
//
// The metrics are retrieved with the given context and metrics list.
func (s *MemStorage) GetMetrics(ctx context.Context, m mtr.MetricsList) (mtr.MetricsList, error) {

	uniqID := make(map[string]bool)
	for _, v := range m {
		if v.MType == GaugeName {
			uniqID[v.ID] = true
		}
		if v.MType == CounterName {
			uniqID[v.ID] = false
		}
	}

	metrics := make(mtr.MetricsList, 0, len(m))

	keys := make([]string, 0, len(uniqID))

	for k := range uniqID {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, id := range keys {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context canceled during processing: %w", ctx.Err())
		default:
			var metric mtr.Metrics
			if uniqID[id] {
				metric = mtr.Metrics{
					ID:    id,
					MType: GaugeName,
					Value: new(float64),
				}
				*metric.Value = s.Gauge[id]
			} else {
				metric = mtr.Metrics{
					ID:    id,
					MType: CounterName,
					Delta: new(int64),
				}
				*metric.Delta = s.Counter[id]
			}
			metrics = append(metrics, metric)
		}
	}
	return metrics, nil
}

// Get gets a metric from the MemStorage instance.
//
// The metric is retrieved with the given type and name.
func (s *MemStorage) Get(metricType string, metricName string) (string, bool) {
	switch metricType {
	case CounterName:
		res, ok := s.Counter[metricName]
		if !ok {
			return "", false
		} else {
			m := strconv.FormatInt(res, 10)
			return m, true
		}
	case GaugeName:
		res, ok := s.Gauge[metricName]
		if !ok {
			return "", false
		} else {
			m := strconv.FormatFloat(res, 'f', -1, 64)
			return m, true
		}
	default:
		return "", false
	}
}

// String returns a string representation of the MemStorage instance.
func (s *MemStorage) String() string {
	var res = ""
	for metricName, metricValue := range s.Gauge {
		res = res + metricName + "=" + strconv.FormatFloat(metricValue, 'f', -1, 64) + "\n"
	}
	for metricName, metricValue := range s.Counter {
		res = res + metricName + "=" + strconv.FormatInt(metricValue, 10) + "\n"
	}
	return res
}

// SaveToFile saves data to a file.
//
// The data is saved to the given file path.
func (s MemStorage) SaveToFile(path string) error {
	data, err := easyjson.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// RestoreFromFile restores data from a file.
//
// The data is restored from the given file path.
func (s *MemStorage) RestoreFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return easyjson.Unmarshal(data, s)
}
