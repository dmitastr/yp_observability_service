package common

import "github.com/caarlos0/env/v6"

type Config struct {
	Address        *string `env:"ADDRESS"`
	PollInterval   *int    `env:"POLL_INTERVAL"`
	ReportInterval *int    `env:"REPORT_INTERVAL"`
}

func New(address string, pollInterval int, reportInterval int) (cfg *Config) {
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
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
