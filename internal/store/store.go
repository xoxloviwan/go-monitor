package store

import (
	"errors"
	"strconv"
)

const counterName = "counter"
const gaugeName = "gauge"

type Gauge map[string]float64

type Counter map[string]int64

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
	case counterName:
		res64, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return err
		}
		s.Counter[metricName] += res64

	case gaugeName:
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
	case counterName:
		res, ok := s.Counter[metricName]
		if !ok {
			return "", false
		} else {
			m := strconv.FormatInt(res, 10)
			return m, true
		}
	case gaugeName:
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

func (s *MemStorage) GetUrls() []string {
	var urls []string
	for metricName, metricValue := range s.Gauge {
		url := "/update/" + gaugeName + "/" + metricName + "/" + strconv.FormatFloat(metricValue, 'f', -1, 64)
		urls = append(urls, url)
	}
	for metricName, metricValue := range s.Counter {
		url := "/update/" + counterName + "/" + metricName + "/" + strconv.FormatInt(metricValue, 10)
		urls = append(urls, url)
	}
	return urls
}
