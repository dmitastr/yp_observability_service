package agentenvconfig

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Address        *string `env:"ADDRESS" mapstructure:"address" json:"address"`
	PollInterval   *int    `env:"POLL_INTERVAL" mapstructure:"poll_interval" json:"poll_interval"`
	ReportInterval *int    `env:"REPORT_INTERVAL" mapstructure:"report_interval" json:"report_interval"`
	Key            *string `env:"KEY" mapstructure:"k"`
	RateLimit      *int    `env:"RATE_LIMIT" mapstructure:"rate_limit" json:"rate_limit"`
	PublicKeyFile  *string `env:"CRYPTO_KEY" mapstructure:"crypto-key" json:"crypto_key"`
	GRPCEnable     *bool   `env:"GRPC_ENABLE" mapstructure:"grpc-enable" json:"grpc-enable"`
}

func NewConfig() (*Config, error) {
	flagSet := pflag.NewFlagSet("agent", pflag.ExitOnError)
	// Unmarshal the configuration into the Config struct
	flagSet.StringP("address", "a", "localhost:8080", "set app host and port")
	flagSet.IntP("report_interval", "r", 10, "frequency of data sending to app in seconds")
	flagSet.IntP("poll_interval", "p", 10, "frequency of metric polling from source in seconds")
	flagSet.IntP("rate_limit", "l", 3, "rate limit")
	flagSet.String("k", "", "key for request signing")
	flagSet.String("crypto-key", "", "path to file with public key")
	flagSet.StringP("config", "c", "", "path to config file")
	flagSet.BoolP("grpc-enable", "g", false, "use gRPC client for sending requests")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	_ = viper.BindPFlags(flagSet)

	viper.AutomaticEnv()

	// Bind environment variables
	_ = viper.BindEnv("address", "ADDRESS")
	_ = viper.BindEnv("k", "KEY")
	_ = viper.BindEnv("report_interval", "REPORT_INTERVAL")
	_ = viper.BindEnv("poll_interval", "POLL_INTERVAL")
	_ = viper.BindEnv("rate_limit", "RATE_LIMIT")
	_ = viper.BindEnv("crypto-key", "CRYPTO_KEY")
	_ = viper.BindEnv("config", "CONFIG")
	_ = viper.BindEnv("grpc-enable", "GRPC_ENABLE")

	if cfgPath := viper.GetString("config"); cfgPath != "" {
		viper.SetConfigFile(cfgPath)
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode metricsAgent config: %w", err)
	}

	return &cfg, nil
}
