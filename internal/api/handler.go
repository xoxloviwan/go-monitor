package api

import (
	"net/http"
)

type Handler struct {
	store ReaderWriter
}

type Reader interface {
	Get(metricType string, metricName string) (string, bool)
}

type Writer interface {
	Add(metricType string, metricName string, metricValue string) error
}

type ReaderWriter interface {
	Reader
	Writer
}

func NewHandler(store ReaderWriter) *http.ServeMux {

	handler := &Handler{
		store: store,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handler.update)
	mux.HandleFunc("/value/{metricType}/{metricName}", handler.value)
	return mux
}

func (hdl *Handler) update(w http.ResponseWriter, req *http.Request) {
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

	err := hdl.store.Add(metricType, metricName, metricValue)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (hdl *Handler) value(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricType := req.PathValue("metricType")
	metricName := req.PathValue("metricName")

	v, ok := hdl.store.Get(metricType, metricName)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(v))
	}

}
