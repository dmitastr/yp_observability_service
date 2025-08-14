package server

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/logger"

	memstorage "github.com/dmitastr/yp_observability_service/internal/datasources/database"
	filestorage "github.com/dmitastr/yp_observability_service/internal/datasources/file_storage"
	postgresstorage "github.com/dmitastr/yp_observability_service/internal/datasources/postgres_storage"
	postgrespinger "github.com/dmitastr/yp_observability_service/internal/domain/postgres_pinger"
	dbinterface "github.com/dmitastr/yp_observability_service/internal/repository/database"

	"github.com/dmitastr/yp_observability_service/internal/domain/service"

	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/get_metric"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/list_metric"
	pingdatabase "github.com/dmitastr/yp_observability_service/internal/presentation/handlers/ping_database"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/update_metric"
	updatemetricsbatch "github.com/dmitastr/yp_observability_service/internal/presentation/handlers/update_metrics_batch"

	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/compress"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/hashsign"
	requestlogger "github.com/dmitastr/yp_observability_service/internal/presentation/middleware/request_logger"
)

func NewServer(cfg serverenvconfig.Config) (*chi.Mux, dbinterface.Database) {
	var storage dbinterface.Database
	if cfg.DBUrl == nil || *cfg.DBUrl == "" {
		fileStorage := filestorage.New(cfg)
		storage = memstorage.NewStorage(cfg, fileStorage)

	} else {
		var err error
		storage, err = postgresstorage.NewPG(context.TODO(), cfg)
		if err != nil {
			logger.GetLogger().Panicf("couldn't connect to postgres database: %v", err)
		}
	}
	if err := storage.Init(); err != nil {
		logger.GetLogger().Panicf("error while initializing database: %v", err)
	}

	pinger := postgrespinger.New()

	observabilityService := service.NewService(storage, pinger)

	metricHandler := updatemetric.NewHandler(observabilityService)
	metricBatchHandler := updatemetricsbatch.NewHandler(observabilityService)
	getMetricHandler := getmetric.NewHandler(observabilityService)
	listMetricsHandler := listmetric.NewHandler(observabilityService)
	pingHandler := pingdatabase.New(observabilityService)

	router := chi.NewRouter()

	// middleware
	hashGenerator := hashsign.NewHashGenerator(cfg.Key)
	router.Use(
		requestlogger.RequestLogger,
		compress.Handler,
		hashGenerator.CheckHash,
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
