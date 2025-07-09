package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/dmitastr/yp_observability_service/internal/server"
)

var serverAddress string

func init() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "set server host and port")
}

func main() {
	flag.Parse()

	router := server.NewServer()

	fmt.Printf("Starting server=%s\n", serverAddress)
	if err := http.ListenAndServe(serverAddress, router); err != nil {
		panic(err)
	}
}
