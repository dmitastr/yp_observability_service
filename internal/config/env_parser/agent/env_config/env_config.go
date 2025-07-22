package common

import (
	"github.com/caarlos0/env/v6"
	"github.com/dmitastr/yp_observability_service/internal/logger"
)

type Config struct {
	Address        *string `env:"ADDRESS"`
	PollInterval   *int    `env:"POLL_INTERVAL"`
	ReportInterval *int    `env:"REPORT_INTERVAL"`
}

func New(address string, pollInterval int, reportInterval int) (cfg Config) {
	err := env.Parse(&cfg)
	if err != nil {
		logger.GetLogger().Errorf("error while reading env variables=%s", err)
	}
	if cfg.Address == nil {
		cfg.Address = &address
	}
	if cfg.ReportInterval == nil {
		cfg.ReportInterval = &reportInterval
	}
	if cfg.PollInterval == nil {
		cfg.PollInterval = &pollInterval
	}
	return
}
