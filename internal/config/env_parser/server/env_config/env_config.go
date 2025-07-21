package common

import "github.com/caarlos0/env/v6"

type Config struct {
	Address         *string `env:"ADDRESS"`
	StoreInterval   *int    `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
}

func New(address string, storeInterval int, fileStoragePath string, restore bool) (cfg Config) {
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}
	if cfg.Address == nil {
		cfg.Address = &address
	}
	if cfg.StoreInterval == nil {
		cfg.StoreInterval = &storeInterval
	}
	if cfg.FileStoragePath == nil {
		cfg.FileStoragePath = &fileStoragePath
	}
	if cfg.Restore == nil {
		cfg.Restore = &restore
	}
	return
}
