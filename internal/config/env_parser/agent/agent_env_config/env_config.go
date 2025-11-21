package agentenvconfig

type Config struct {
	Address        *string `env:"ADDRESS" mapstructure:"address" json:"address"`
	PollInterval   *int    `env:"POLL_INTERVAL" mapstructure:"poll_interval" json:"poll_interval"`
	ReportInterval *int    `env:"REPORT_INTERVAL" mapstructure:"report_interval" json:"report_interval"`
	Key            *string `env:"KEY" mapstructure:"k"`
	RateLimit      *int    `env:"RATE_LIMIT" mapstructure:"rate_limit" json:"rate_limit"`
	PublicKeyFile  *string `env:"CRYPTO_KEY" mapstructure:"crypto-key" json:"crypto_key"`
	GRPCEnable     *bool   `env:"GRPC_ENABLE" mapstructure:"grpc-enable" json:"grpc-enable"`
}
