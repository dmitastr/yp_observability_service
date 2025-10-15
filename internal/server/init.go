package server

import (
	"context"
	"net/http"

	"github.com/dmitastr/yp_observability_service/internal/domain/audit/listener"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dmitastr/yp_observability_service/internal/domain/audit"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/hash"
	"github.com/go-chi/chi/v5"

	"github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	db "github.com/dmitastr/yp_observability_service/internal/datasources/database"
	filestorage "github.com/dmitastr/yp_observability_service/internal/datasources/file_storage"
	postgresstorage "github.com/dmitastr/yp_observability_service/internal/datasources/postgres_storage"
	postgrespinger "github.com/dmitastr/yp_observability_service/internal/domain/postgres_pinger"
	"github.com/dmitastr/yp_observability_service/internal/domain/service"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/get_metric"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/list_metric"
	pingdatabase "github.com/dmitastr/yp_observability_service/internal/presentation/handlers/ping_database"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/update_metric"
	updatemetricsbatch "github.com/dmitastr/yp_observability_service/internal/presentation/handlers/update_metrics_batch"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/compress"
	requestlogger "github.com/dmitastr/yp_observability_service/internal/presentation/middleware/request_logger"
	dbinterface "github.com/dmitastr/yp_observability_service/internal/repository/database"
)

func NewServer(cfg serverenvconfig.Config) (*chi.Mux, dbinterface.Database) {
	var storage dbinterface.Database
	if cfg.DBUrl == nil || *cfg.DBUrl == "" {
		fileStorage := filestorage.New(cfg)
		storage = db.NewStorage(cfg, fileStorage)

	} else {
		var err error
		storage, err = postgresstorage.NewPG(context.TODO(), cfg)
		if err != nil {
			logger.GetLogger().Panicf("couldn't connect to postgres database: ", err)
		}
	}
	storage.Init()

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

	router := chi.NewRouter()

	// middleware
	router.Use(
		requestlogger.RequestLogger,
		signedCheckHandler.Check,
		compress.CompressMiddleware,
	)

	// setting routes
	router.Get(`/`, listMetricsHandler.ServeHTTP)

	router.Route(`/update`, func(r chi.Router) {
		r.Post(`/`, metricHandler.ServeHTTP)
		r.Post(`/{mtype}/{name}/{value}`, metricHandler.ServeHTTP)
	})

	router.Post(`/updates/`, metricBatchHandler.ServeHTTP)

	router.Route(`/value`, func(r chi.Router) {
		r.Post(`/`, getMetricHandler.ServeHTTP)
		r.Get(`/{mtype}/{name}`, getMetricHandler.ServeHTTP)
	})

	router.Get(`/ping`, pingHandler.ServeHTTP)

	return router, storage
}

func Execute() error {
	rootCmd := &cobra.Command{
		Use: "YP observability service",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Initialize()

			var cfg serverenvconfig.Config
			// Unmarshal the configuration into the Config struct
			if err := viper.Unmarshal(&cfg); err != nil {
				logger.GetLogger().Errorf("Unable to decode into struct, %v\n", err)
				return err
			}
			router, postgresDB := NewServer(cfg)
			defer postgresDB.Close()

			logger.GetLogger().Infof("Starting server=%s, store interval=%d, file storage path=%s, restore data=%t\n", *cfg.Address, *cfg.StoreInterval, *cfg.FileStoragePath, *cfg.Restore)
			if err := http.ListenAndServe(*cfg.Address, router); err != nil {
				return err
			}

			return nil
		},
	}

	rootCmd.Flags().String("a", "localhost:8080", "set server host and port")
	rootCmd.Flags().Int("i", 300, "interval for storing data to the file in seconds, 0=stream writing")
	rootCmd.Flags().Bool("r", false, "restore data from file")
	rootCmd.Flags().String("f", "./data/data.json", "path for writing data")
	rootCmd.Flags().String("d", "", "database connection url")
	rootCmd.Flags().String("k", "", "key for request signing")
	rootCmd.Flags().String("audit-file", "", "file path for audit logs")
	rootCmd.Flags().String("audit-url", "", "url for audit logs")

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

	return rootCmd.Execute()

}
