package common

import "github.com/caarlos0/env/v6"

type Config struct {
	Address *string `env:"ADDRESS"`
}

func New(address string) (cfg Config) {
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}
	if cfg.Address == nil {
		cfg.Address = &address
	}
	return
}
