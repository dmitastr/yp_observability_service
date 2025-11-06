package serverenvconfig

type Config struct {
	Address         *string `env:"ADDRESS" mapstructure:"a"`
	StoreInterval   *int    `env:"STORE_INTERVAL" mapstructure:"i"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH" mapstructure:"f"`
	Restore         *bool   `env:"RESTORE" mapstructure:"r"`
	DBUrl           *string `env:"DATABASE_DSN" mapstructure:"d"`
	Key             *string `env:"KEY" mapstructure:"k"`
	AuditFile       *string `env:"AUDIT_FILE" mapstructure:"audit-file"`
	AuditURL        *string `env:"AUDIT_URL" mapstructure:"audit-url"`
	PrivateKeyPath  *string `env:"CRYPTO_KEY" mapstructure:"crypto-key"`
}
