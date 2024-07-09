package api

import (
	"net/http"

	"github.com/xoxloviwan/go-monitor/internal/store"
)

func NewHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", update)
	mux.HandleFunc("/value/{metricType}/{metricName}", value)
	return mux
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

	err := store.Storage.Add(metricType, metricName, metricValue)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func value(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricType := req.PathValue("metricType")
	metricName := req.PathValue("metricName")

	v, ok := store.Storage.Get(metricType, metricName)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Write([]byte(v))
	}

}
