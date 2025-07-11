package main

import (
	"flag"
	"fmt"
	"net/http"

	envconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/env_config"
	"github.com/dmitastr/yp_observability_service/internal/server"
)


var serverAddress string

func init() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "set server host and port")
}

func main() {
	flag.Parse()
	cfg := envconfig.New(serverAddress)

	router := server.NewServer()

	fmt.Printf("Starting server=%s\n", *cfg.Address)
	if err := http.ListenAndServe(*cfg.Address, router); err != nil {
		panic(err)
	}
}
