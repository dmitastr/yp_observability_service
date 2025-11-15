package serverenvconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Address         *string `env:"ADDRESS" mapstructure:"address"`
	StoreInterval   *int    `env:"STORE_INTERVAL" mapstructure:"store_interval"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH" mapstructure:"store_file"`
	Restore         *bool   `env:"RESTORE" mapstructure:"restore"`
	DBUrl           *string `env:"DATABASE_DSN" mapstructure:"database_dsn"`
	Key             *string `env:"KEY" mapstructure:"k"`
	AuditFile       *string `env:"AUDIT_FILE" mapstructure:"audit-file"`
	AuditURL        *string `env:"AUDIT_URL" mapstructure:"audit-url"`
	PrivateKeyPath  *string `env:"CRYPTO_KEY" mapstructure:"crypto-key"`
	TrustedSubnet   *string `env:"TRUSTED_SUBNET" mapstructure:"trusted_subnet"`
}

// New reads command line and env arguments, reads config file if any
// and creates [Config] instance
func New() (*Config, error) {
	flagSet := pflag.NewFlagSet("observability", pflag.ExitOnError)
	flagSet.StringP("address", "a", "localhost:8080", "set app host and port")
	flagSet.IntP("store_interval", "i", 300, "interval for storing data to the file in seconds, 0=stream writing")
	flagSet.BoolP("restore", "r", false, "restore data from file")
	flagSet.StringP("store_file", "f", "./data/data.json", "path for writing data")
	flagSet.StringP("database_dsn", "d", "", "postgres connection url")
	flagSet.StringP("key", "k", "", "key for request signing")
	flagSet.String("audit-file", "", "file path for audit logs")
	flagSet.String("audit-url", "", "url for audit logs")
	flagSet.String("crypto-key", "", "path to file with private key")
	flagSet.StringP("config", "c", "", "path to config file")
	flagSet.StringP("trusted_subnet", "t", "", "trusted subnet in CIDR format")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	_ = viper.BindPFlags(flagSet)

	viper.AutomaticEnv()

	// Bind environment variables
	_ = viper.BindEnv("a", "ADDRESS")
	_ = viper.BindEnv("i", "STORE_INTERVAL")
	_ = viper.BindEnv("f", "FILE_STORAGE_PATH")
	_ = viper.BindEnv("r", "RESTORE")
	_ = viper.BindEnv("d", "DATABASE_DSN")
	_ = viper.BindEnv("k", "KEY")
	_ = viper.BindEnv("audit-file", "AUDIT_FILE")
	_ = viper.BindEnv("audit-url", "AUDIT_URL")
	_ = viper.BindEnv("crypto-key", "CRYPTO_KEY")
	_ = viper.BindEnv("config", "CONFIG")
	_ = viper.BindEnv("trusted_subnet", "TRUSTED_SUBNET")

	if cfgPath := viper.GetString("config"); cfgPath != "" {
		viper.SetConfigFile(cfgPath)
		if err := viper.ReadInConfig(); err != nil {
			logger.Errorf("error reading config file, %s\n", err)
		}
	}

	var cfg Config
	// Unmarshal the configuration into the Config struct
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %w", err)
	}
	return &cfg, nil

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
