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

var Storage MemStorage = MemStorage{
	gauge:   make(map[string]float64),
	counter: make(map[string]int64),
}

func (s *MemStorage) Add(metricType string, metricName string, metricValue string) (err error) {
	switch metricType {
	case "counter":
		res64, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return err
		}
		Storage.counter[metricName] = Storage.counter[metricName] + res64
	case "gauge":
		res64, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return err
		}
		Storage.gauge[metricName] = res64
	default:
		return errors.New("unknown metric type")
	}
	return err
}

func (s *MemStorage) Get(metricType string, metricName string) (string, bool) {
	switch metricType {
	case "counter":
		res, ok := Storage.counter[metricName]
		if !ok {
			return "", false
		} else {
			m := strconv.FormatInt(res, 10)
			return m, true
		}
	case "gauge":
		res, ok := Storage.gauge[metricName]
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
