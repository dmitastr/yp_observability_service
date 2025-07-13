package main

import (
	"flag"

	"github.com/dmitastr/yp_observability_service/internal/agent/client"
	envconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/agent/env_config"
	"github.com/dmitastr/yp_observability_service/internal/logger"
)


var serverAddress string
var reportInterval int
var pollInterval int

func init() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "set server host and port")
	flag.IntVar(&reportInterval, "r", 10, "frequency of data sending to server in seconds")
	flag.IntVar(&pollInterval, "p", 2, "frequency of metric polling from source in seconds")
}

func main() {
	flag.Parse()
	logger.Initialize()
	cfg := envconfig.New(serverAddress, pollInterval, reportInterval)

	logger.GetLogger().Infof("Starting client for server=%s, poll interval=%d, report interval=%d", 
		*cfg.Address, 
		*cfg.PollInterval, 
		*cfg.ReportInterval,
	)

	agent := client.NewAgent(*cfg.Address)
	agent.Run(*cfg.PollInterval, *cfg.ReportInterval)
}
