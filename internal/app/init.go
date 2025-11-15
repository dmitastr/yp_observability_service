package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/dmitastr/yp_observability_service/internal/domain/audit"
	"github.com/dmitastr/yp_observability_service/internal/domain/audit/listener"
	"github.com/dmitastr/yp_observability_service/internal/domain/pinger/postgres_pinger"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/certdecode"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/hash"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/ipchecker"
	dbinterface "github.com/dmitastr/yp_observability_service/internal/repository"
	"github.com/dmitastr/yp_observability_service/internal/repository/filestorage"
	db "github.com/dmitastr/yp_observability_service/internal/repository/memstorage"
	"github.com/dmitastr/yp_observability_service/internal/repository/postgres_storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/domain/service"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/get_metric"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/list_metric"
	pingdatabase "github.com/dmitastr/yp_observability_service/internal/presentation/handlers/ping_database"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/update_metric"
	updatemetricsbatch "github.com/dmitastr/yp_observability_service/internal/presentation/handlers/update_metrics_batch"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/compress"
	requestlogger "github.com/dmitastr/yp_observability_service/internal/presentation/middleware/request_logger"
)

// NewApp creates a new app, register all handlers and middleware
// and inject necessary dependencies
func NewApp(ctx context.Context) (*http.Server, dbinterface.Database, error) {
	cfg, err := serverenvconfig.New()
	if err != nil {
		return nil, nil, err
	}

	var storage dbinterface.Database
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
	ipValidator, err := ipchecker.New(*cfg.TrustedSubnet)
	if err != nil {
		return nil, storage, fmt.Errorf("error creating ip validator: %w", err)
	}

	// middleware
	router.Use(
		ipValidator.Handle,
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
	server := &http.Server{
		Addr:              *cfg.Address,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		Handler:           router,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}

	return server, storage, nil
}
