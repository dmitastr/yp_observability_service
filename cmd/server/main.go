package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/caarlos0/env/v6"

	"github.com/dmitastr/yp_observability_service/internal/common"
	"github.com/dmitastr/yp_observability_service/internal/server"
)


var serverAddress string
var cfg common.Config

func init() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "set server host and port")

	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}
	if cfg.Address == nil {
		cfg.Address = &serverAddress
	}

}

func main() {
	flag.Parse()

	router := server.NewServer()

	fmt.Printf("Starting server=%s\n", *cfg.Address)
	if err := http.ListenAndServe(*cfg.Address, router); err != nil {
		panic(err)
	}
}
