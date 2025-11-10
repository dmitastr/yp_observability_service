package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/dmitastr/yp_observability_service/internal/domain/audit/listener"
	"github.com/dmitastr/yp_observability_service/internal/domain/pinger/postgres_pinger"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/certdecode"
	dbinterface "github.com/dmitastr/yp_observability_service/internal/repository"
	"github.com/dmitastr/yp_observability_service/internal/repository/filestorage"
	db "github.com/dmitastr/yp_observability_service/internal/repository/memstorage"
	"github.com/dmitastr/yp_observability_service/internal/repository/postgres_storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/dmitastr/yp_observability_service/internal/domain/audit"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/hash"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/domain/service"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/get_metric"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/list_metric"
	pingdatabase "github.com/dmitastr/yp_observability_service/internal/presentation/handlers/ping_database"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/update_metric"
	updatemetricsbatch "github.com/dmitastr/yp_observability_service/internal/presentation/handlers/update_metrics_batch"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/compress"
	requestlogger "github.com/dmitastr/yp_observability_service/internal/presentation/middleware/request_logger"
)

// NewServer creates a new server, register all handlers and middleware
// and inject necessary dependencies
func NewServer(ctx context.Context, cfg serverenvconfig.Config) (server *http.Server, storage dbinterface.Database, err error) {
	if cfg.DBUrl == nil || *cfg.DBUrl == "" {
		fileStorage := filestorage.New(cfg)
		storage = db.NewStorage(cfg, fileStorage)

	} else {
		storage, err = postgresstorage.NewPG(ctx, cfg)
		if err != nil {
			return nil, storage, fmt.Errorf("error creating postgres storage: %w", err)
		}
	}
	if err = storage.Init("file://migrations"); err != nil {
		return nil, storage, fmt.Errorf("error initializing postgres storage: %w", err)
	}

	router := chi.NewRouter()

	pinger := postgrespinger.New()
	auditor := audit.NewAuditor().
		AddListener(listener.NewListener(listener.FileListenerType, cfg.AuditFile)).
		AddListener(listener.NewListener(listener.URLListenerType, cfg.AuditURL))

	observabilityService := service.NewService(storage, pinger, auditor)

	metricHandler := updatemetric.NewHandler(observabilityService)
	metricBatchHandler := updatemetricsbatch.NewHandler(observabilityService)
	getMetricHandler := getmetric.NewHandler(observabilityService)
	listMetricsHandler := listmetric.NewHandler(observabilityService)
	pingHandler := pingdatabase.New(observabilityService)
	signedCheckHandler := hash.NewSignedChecker(cfg)
	rsaDecodeHandler := certdecode.NewCertDecoder(*cfg.PrivateKeyPath)

	// middleware
	// gracefulShutdownHandler := gracefulshutdown.NewGracefulShutdownHandler(cancelCh)
	router.Use(
		requestlogger.Handle,
		rsaDecodeHandler.Handle,
		signedCheckHandler.Handle,
	)

	// Set path for profiling
	router.Mount("/debug", middleware.Profiler())

	// setting routes
	router.Group(func(r chi.Router) {
		r.Use(compress.HandleCompression)
		r.Get(`/`, listMetricsHandler.ServeHTTP)

		r.Route(`/update`, func(r chi.Router) {
			r.Post(`/`, metricHandler.ServeHTTP)
			r.Post(`/{mtype}/{name}/{value}`, metricHandler.ServeHTTP)
		})

		r.Post(`/updates/`, metricBatchHandler.ServeHTTP)
		r.Get(`/ping`, pingHandler.ServeHTTP)

	})

	router.Group(func(r chi.Router) {
		r.Use(compress.HandleDecompression)

		r.Route(`/value`, func(r chi.Router) {
			r.Post(`/`, getMetricHandler.ServeHTTP)
			r.Get(`/{mtype}/{name}`, getMetricHandler.ServeHTTP)
		})

	})
	server = &http.Server{
		Addr:              *cfg.Address,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		Handler:           router,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}

	return
}

// Run initialized [cobra.Command] for args parsing and starts the server
func Run(ctx context.Context) error {
	rootCmd := &cobra.Command{
		Use: "YP observability service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := logger.Initialize(); err != nil {
				return err
			}
			if cfgPath := viper.GetString("config"); cfgPath != "" {
				viper.SetConfigFile(cfgPath)
				if err := viper.ReadInConfig(); err != nil {
					logger.Errorf("Error reading config file, %s\n", err)
				}
			}

			var cfg serverenvconfig.Config
			// Unmarshal the configuration into the Config struct
			if err := viper.Unmarshal(&cfg); err != nil {
				logger.Errorf("Unable to decode into struct, %v\n", err)
				return err
			}
			server, postgresDB, err := NewServer(ctx, cfg)
			if err != nil {
				return err
			}

			g, gCtx := errgroup.WithContext(ctx)
			// Server goroutine
			g.Go(func() error {
				logger.Infof("Starting server with config: %s\n", cfg.String())
				if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					return fmt.Errorf("server error: %w", err)
				}
				logger.Info("Server stopped")
				return nil
			})

			// Server shutdown goroutine
			g.Go(func() error {
				<-gCtx.Done()
				shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				return server.Shutdown(shutdownCtx)
			})

			// Database closing goroutine
			g.Go(func() error {
				<-gCtx.Done()
				return postgresDB.Close()
			})

			if err := g.Wait(); err != nil {
				logger.Infof("exit reason: %v", err)
			}

			return nil
		},
	}

	rootCmd.Flags().StringP("address", "a", "localhost:8080", "set server host and port")
	rootCmd.Flags().IntP("store_interval", "i", 300, "interval for storing data to the file in seconds, 0=stream writing")
	rootCmd.Flags().BoolP("restore", "r", false, "restore data from file")
	rootCmd.Flags().StringP("store_file", "f", "./data/data.json", "path for writing data")
	rootCmd.Flags().StringP("database_dsn", "d", "", "postgres connection url")
	rootCmd.Flags().StringP("key", "k", "", "key for request signing")
	rootCmd.Flags().String("audit-file", "", "file path for audit logs")
	rootCmd.Flags().String("audit-url", "", "url for audit logs")
	rootCmd.Flags().String("crypto-key", "", "path to file with private key")
	rootCmd.Flags().StringP("config", "c", "", "path to config file")

	_ = viper.BindPFlags(rootCmd.Flags())

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

	return rootCmd.Execute()

}
