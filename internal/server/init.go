package server

import (
	"github.com/go-chi/chi/v5"

	envconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/env_config"
	db "github.com/dmitastr/yp_observability_service/internal/datasources/database"
	filestorage "github.com/dmitastr/yp_observability_service/internal/datasources/file_storage"
	"github.com/dmitastr/yp_observability_service/internal/domain/service"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/get_metric"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/list_metric"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/update_metric"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/compress"
	requestlogger "github.com/dmitastr/yp_observability_service/internal/presentation/middleware/request_logger"
)

func NewServer(cfg envconfig.Config) *chi.Mux {
	storage := db.NewStorage(*cfg.FileStoragePath, *cfg.StoreInterval, *cfg.Restore)
	service := service.NewService(storage)

	if *cfg.StoreInterval != 0 {
		fileStorage := filestorage.New(storage, *cfg.StoreInterval)
		fileStorage.Run()
	}

	metricHandler := updatemetric.NewHandler(service)
	getMetricHandler := getmetric.NewHandler(service)
	listMetricsHandler := listmetric.NewHandler(service)

	router := chi.NewRouter()

	// middleware
	router.Use(
		requestlogger.RequestLogger, 
		compress.CompressMiddleware,
	)

	// setting routes
	router.Get(`/`, listMetricsHandler.ServeHTTP)
	
	router.Route(`/update`, func(r chi.Router) {
		r.Post(`/`, metricHandler.ServeHTTP)
		r.Post(`/{mtype}/{name}/{value}`, metricHandler.ServeHTTP)
	})

	router.Route(`/value`, func(r chi.Router) {
		r.Post(`/`, getMetricHandler.ServeHTTP)
		r.Get(`/{mtype}/{name}`, getMetricHandler.ServeHTTP)
	})
	
	return router
}
