package main

import (
	"net/http"

	"github.com/xoxloviwan/go-monitor/internal/api"
)

func main() {
	handler := api.NewHandler()
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		panic(err)
	}
}
