package api

import (
	"net/http"
)

func NewHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", update)
	mux.HandleFunc("/value/{metricType}/{metricName}", value)
	return mux
}
