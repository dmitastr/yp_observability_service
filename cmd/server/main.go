package main

import (
	"net/http"

	db "github.com/dmitastr/yp_observability_service/internal/database"
	"github.com/dmitastr/yp_observability_service/internal/domain/service"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/update_metric"
	"github.com/dmitastr/yp_observability_service/internal/presentation/handlers/get_metric"
)


func main() {
	storage := db.NewStorage()
	service := service.NewService(storage)
	metricHandler := updatemetric.NewHandler(service)
	getMetricHandler := getmetric.NewHandler(service)

	mux := http.NewServeMux()
	mux.Handle(`POST /update/{mtype}/{name}/{value}`, *metricHandler)
	mux.Handle(`GET /value/{mtype}/{name}`, *getMetricHandler)
	
	if err := http.ListenAndServe(`:8080`, mux); err != nil {
		panic(err)
	}
}