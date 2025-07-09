package main

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/dmitastr/yp_observability_service/internal/agent/client"
)

type Config struct {
	Address        *string `env:"ADDRESS"`
	PollInterval   *int    `env:"POLL_INTERVAL"`
	ReportInterval *int    `env:"REPORT_INTERVAL"`
}

var serverAddress string
var reportInterval int
var pollInterval int
var cfg Config

func init() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "set server host and port")
	flag.IntVar(&reportInterval, "r", 10, "frequency of data sending to server in seconds")
	flag.IntVar(&pollInterval, "p", 2, "frequency of metric polling from source in seconds")

	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	if cfg.Address == nil {
		cfg.Address = &serverAddress
	}
	if cfg.ReportInterval == nil {
		cfg.ReportInterval = &reportInterval
	}
	if cfg.PollInterval == nil {
		cfg.PollInterval = &pollInterval
	}

}

func main() {
	flag.Parse()
	fmt.Printf("Starting client for server=%s, poll interval=%d, report interval=%d\n", *cfg.Address, *cfg.PollInterval, *cfg.ReportInterval)

	agent := client.NewAgent(*cfg.Address)
	agent.Run(*cfg.PollInterval, *cfg.ReportInterval)
}
