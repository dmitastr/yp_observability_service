package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"net/http"

	"github.com/dmitastr/yp_observability_service/internal/server"
)

type Config struct {
	Address        string `env:"ADDRESS"`
}

var serverAddress string
var cfg Config

func init() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "set server host and port")

	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}
	if cfg.Address == "" {
		cfg.Address = serverAddress
	}

}

func main() {
	flag.Parse()

	router := server.NewServer()

	fmt.Printf("Starting server=%s\n", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, router); err != nil {
		panic(err)
	}
}
