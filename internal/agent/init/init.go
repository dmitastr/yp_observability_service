package init

import (
	"fmt"

	"context"

	"github.com/dmitastr/yp_observability_service/internal/agent/client"
	config "github.com/dmitastr/yp_observability_service/internal/config/env_parser/agent/agent_env_config"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Run initialized [cobra.Command] for args parsing and starts the app
func Run(ctx context.Context) error {
	rootCmd := &cobra.Command{
		Use: "YP observability agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := logger.Initialize(); err != nil {
				return err
			}

			if cfgPath := viper.GetString("config"); cfgPath != "" {
				viper.SetConfigFile(cfgPath)
				if err := viper.ReadInConfig(); err != nil {
					return fmt.Errorf("error reading config file: %w\n", err)
				}
			}

			var cfg config.Config
			// Unmarshal the configuration into the Config struct
			if err := viper.Unmarshal(&cfg); err != nil {
				return fmt.Errorf("unable to decode agent config: %w\n", err)
			}
			agent, err := client.NewAgent(cfg)
			if err != nil {
				return fmt.Errorf("error initializing agent: %w\n", err)
			}

			logger.Infof("Starting client for app=%s, poll interval=%d, report interval=%d",
				*cfg.Address,
				*cfg.PollInterval,
				*cfg.ReportInterval,
			)

			return agent.Run(ctx, *cfg.PollInterval, *cfg.ReportInterval)
		},
	}

	rootCmd.Flags().StringP("address", "a", "localhost:8080", "set app host and port")
	rootCmd.Flags().IntP("report_interval", "r", 10, "frequency of data sending to app in seconds")
	rootCmd.Flags().IntP("poll_interval", "p", 10, "frequency of metric polling from source in seconds")
	rootCmd.Flags().IntP("rate_limit", "l", 3, "rate limit")
	rootCmd.Flags().String("k", "", "key for request signing")
	rootCmd.Flags().String("crypto-key", "", "path to file with public key")
	rootCmd.Flags().StringP("config", "c", "", "path to config file")

	_ = viper.BindPFlags(rootCmd.Flags())

	viper.AutomaticEnv()

	// Bind environment variables
	_ = viper.BindEnv("address", "ADDRESS")
	_ = viper.BindEnv("k", "KEY")
	_ = viper.BindEnv("report_interval", "REPORT_INTERVAL")
	_ = viper.BindEnv("poll_interval", "POLL_INTERVAL")
	_ = viper.BindEnv("rate_limit", "RATE_LIMIT")
	_ = viper.BindEnv("crypto-key", "CRYPTO_KEY")
	_ = viper.BindEnv("config", "CONFIG")

	return rootCmd.Execute()

}
