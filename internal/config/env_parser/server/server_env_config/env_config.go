package serverenvconfig

import (
	"github.com/caarlos0/env/v6"
	"github.com/dmitastr/yp_observability_service/internal/logger"
)

type Config struct {
	Address         *string `env:"ADDRESS"`
	StoreInterval   *int    `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
	DBUrl           *string `env:"DB_URL"`
}

func New(address string, storeInterval int, fileStoragePath string, restore bool, dbURL string) (cfg Config) {
	err := env.Parse(&cfg)
	if err != nil {
		logger.GetLogger().Errorf("error while reading env variables=%s", err)
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
	if cfg.DBUrl == nil {
		cfg.DBUrl = &dbURL
	}
	return
}
