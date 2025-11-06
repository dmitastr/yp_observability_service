package agentenvconfig

import (
	"github.com/caarlos0/env/v6"
	"github.com/dmitastr/yp_observability_service/internal/logger"
)

type Config struct {
	Address        *string `env:"ADDRESS" mapstructure:"a"`
	PollInterval   *int    `env:"POLL_INTERVAL" mapstructure:"p"`
	ReportInterval *int    `env:"REPORT_INTERVAL" mapstructure:"r"`
	Key            *string `env:"KEY" mapstructure:"k"`
	RateLimit      *int    `env:"RATE_LIMIT" mapstructure:"l"`
	PublicKeyFile  *string `env:"CRYPTO_KEY" mapstructure:"crypto-key"`
}

func New(address string, pollInterval int, reportInterval int, key string, rateLimit int) (cfg Config) {
	err := env.Parse(&cfg)
	if err != nil {
		logger.Errorf("error while reading env variables=%s", err)
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
	if cfg.Key == nil {
		cfg.Key = &key
	}
	if cfg.RateLimit == nil {
		cfg.RateLimit = &rateLimit
	}
	return
}
