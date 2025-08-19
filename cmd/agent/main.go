package main

import (
	"flag"

	"github.com/dmitastr/yp_observability_service/internal/agent/client"
	"github.com/dmitastr/yp_observability_service/internal/config/env_parser/agent/agent_env_config"
	"github.com/dmitastr/yp_observability_service/internal/logger"
)

var serverAddress string
var reportInterval int
var pollInterval int
var key string
var rateLimit int

func init() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "set server host and port")
	flag.IntVar(&reportInterval, "r", 10, "frequency of data sending to server in seconds")
	flag.IntVar(&pollInterval, "p", 2, "frequency of metric polling from source in seconds")
	flag.StringVar(&key, "k", "", "key for request signing")
	flag.IntVar(&rateLimit, "l", 3, "rate limit")
}

func main() {
	flag.Parse()
	cfg := agentenvconfig.New(serverAddress, pollInterval, reportInterval, key, rateLimit)

	logger.GetLogger().Infof("Starting client for server=%s, poll interval=%d, report interval=%d",
		*cfg.Address,
		*cfg.PollInterval,
		*cfg.ReportInterval,
	)

	agent := client.NewAgent(cfg)
	agent.Run(*cfg.PollInterval, *cfg.ReportInterval)
}
