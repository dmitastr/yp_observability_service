package main

import (
	"net/http"

	db "github.com/dmitastr/yp_observability_service/internal/database"
	"github.com/dmitastr/yp_observability_service/internal/service"
	"github.com/dmitastr/yp_observability_service/internal/handlers/update_metric"
)


func main() {
	storage := db.NewStorage()
	service := service.NewService(storage)
	metricHandler := updatemetric.NewHandler(service)

	mux := http.NewServeMux()
	mux.Handle(`POST /update/{mtype}/{name}/{value}`, *metricHandler)
	
	if err := http.ListenAndServe(`:8080`, mux); err != nil {
		panic(err)
	}
}