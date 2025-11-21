package http

import (
	"context"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/domain/service"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/get_metric"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/list_metric"
	pingdatabase "github.com/dmitastr/yp_observability_service/internal/presentation/handlers/ping_database"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/update_metric"
	updatemetricsbatch "github.com/dmitastr/yp_observability_service/internal/presentation/handlers/update_metrics_batch"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/certdecode"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/compress"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/hash"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/ipchecker"
	requestlogger "github.com/dmitastr/yp_observability_service/internal/presentation/middleware/request_logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewServer creates a new server, register all handlers and middleware
// and inject necessary dependencies
func NewServer(ctx context.Context, cfg *serverenvconfig.Config, iService service.IService, ipValidator *ipchecker.IPValidator) (*http.Server, error) {
	router := chi.NewRouter()

	metricHandler := updatemetric.NewHandler(iService)
	metricBatchHandler := updatemetricsbatch.NewHandler(iService)
	getMetricHandler := getmetric.NewHandler(iService)
	listMetricsHandler := listmetric.NewHandler(iService)
	pingHandler := pingdatabase.New(iService)
	signedCheckHandler := hash.NewSignedChecker(cfg)
	rsaDecodeHandler := certdecode.NewCertDecoder(*cfg.PrivateKeyPath)

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

	return server, nil
}
