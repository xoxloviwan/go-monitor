package api

import (
	"net/http"
	"strconv"
)

func NewHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", update)
	return mux
}

type gauge map[string]float64

type counter map[string]int64

type MemStorage struct {
	gauge
	counter
}

var storage MemStorage = MemStorage{
	gauge:   make(map[string]float64),
	counter: make(map[string]int64),
}

func update(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricType := req.PathValue("metricType")
	metricName := req.PathValue("metricName")
	metricValue := req.PathValue("metricValue")

	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch metricType {
	case "counter":
		res64, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		storage.counter[metricName] = res64 + storage.counter[metricName]
	case "gauge":
		res64, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		storage.gauge[metricName] = res64
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
	//fmt.Printf("%+v\n", storage)
}
