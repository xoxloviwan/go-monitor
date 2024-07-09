package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/xoxloviwan/go-monitor/internal/api"
)

func main() {
	adr := flag.String("a", "localhost:8080", "server adress")
	flag.Parse()
	if len(flag.Args()) > 0 {
		fmt.Println("Too many arguments")
		os.Exit(1)
	}
	r := api.SetupRouter()
	r.Use(gin.Logger())
	err := r.Run(*adr)
	if err != nil {
		panic(err)
	}
}
