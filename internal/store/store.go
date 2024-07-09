package store

import (
	"errors"
	"strconv"
)

type gauge map[string]float64

type counter map[string]int64

type MemStorage struct {
	gauge
	counter
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

func (s *MemStorage) Add(metricType string, metricName string, metricValue string) (err error) {
	switch metricType {
	case "counter":
		res64, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return err
		}
		s.counter[metricName] = s.counter[metricName] + res64
	case "gauge":
		res64, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return err
		}
		s.gauge[metricName] = res64
	default:
		return errors.New("unknown metric type")
	}
	return err
}

func (s *MemStorage) Get(metricType string, metricName string) (string, bool) {
	switch metricType {
	case "counter":
		res, ok := s.counter[metricName]
		if !ok {
			return "", false
		} else {
			m := strconv.FormatInt(res, 10)
			return m, true
		}
	case "gauge":
		res, ok := s.gauge[metricName]
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
