package store

import (
	"errors"
	"os"
	"strconv"

	"github.com/mailru/easyjson"
)

const CounterName = "counter"
const GaugeName = "gauge"

type Gauge map[string]float64

type Counter map[string]int64

// easyjson:json
type MemStorage struct {
	Gauge
	Counter
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

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

func (s *MemStorage) SaveToFile(path string) error {
	data, err := easyjson.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (s *MemStorage) RestoreFromFile(path string) error {

	return nil
}
