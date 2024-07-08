package api

import (
	"net/http"

	"github.com/xoxloviwan/go-monitor/internal/store"
)

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
