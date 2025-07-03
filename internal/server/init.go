package server


import (
	"github.com/go-chi/chi/v5"

	db "github.com/dmitastr/yp_observability_service/internal/database"
	"github.com/dmitastr/yp_observability_service/internal/domain/service"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/get_metric"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/list_metric"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/update_metric"
)

func NewServer() *chi.Mux {
	storage := db.NewStorage()
	service := service.NewService(storage)
	metricHandler := updatemetric.NewHandler(service)
	getMetricHandler := getmetric.NewHandler(service)
	listMetricsHandler := listmetric.NewHandler(service)

	router := chi.NewRouter()

	router.Get(`/`, listMetricsHandler.ServeHTTP)
	router.Post(`/update/{mtype}/{name}/{value}`, metricHandler.ServeHTTP)
	router.Get(`/value/{mtype}/{name}`, getMetricHandler.ServeHTTP)
	return router
}
