package serverenvconfig

import (
	"fmt"
	"strings"
)

type Config struct {
	Address         *string `env:"ADDRESS" mapstructure:"address" json:"address"`
	StoreInterval   *int    `env:"STORE_INTERVAL" mapstructure:"store_interval" json:"store_interval"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH" mapstructure:"store_file"`
	Restore         *bool   `env:"RESTORE" mapstructure:"restore" json:"restore"`
	DBUrl           *string `env:"DATABASE_DSN" mapstructure:"database_dsn" json:"database_dsn"`
	Key             *string `env:"KEY" mapstructure:"k"`
	AuditFile       *string `env:"AUDIT_FILE" mapstructure:"audit-file"`
	AuditURL        *string `env:"AUDIT_URL" mapstructure:"audit-url"`
	PrivateKeyPath  *string `env:"CRYPTO_KEY" mapstructure:"crypto-key" json:"crypto_key"`
}

func (c *Config) String() string {
	s := strings.Builder{}
	if c.Address != nil {
		s.WriteString(fmt.Sprintf("Address:\t%s, ", *c.Address))
	}
	if c.DBUrl != nil {
		s.WriteString(fmt.Sprintf("DBUrl:\t%s, ", *c.DBUrl))
	}
	if c.PrivateKeyPath != nil {
		s.WriteString(fmt.Sprintf("PrivateKeyPath:\t%s, ", *c.PrivateKeyPath))
	}
	return s.String()
}
