package api

import (
	"net/http"

	"github.com/xoxloviwan/go-monitor/internal/store"
)

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
