package pingdatabase

import (
	"context"
	"net/http"
	"time"

	srv "github.com/dmitastr/yp_observability_service/internal/domain/service"
)

// PingDatabaseHandler handles requests for checking database health
type PingDatabaseHandler struct {
	service srv.IService
}

func New(service srv.IService) *PingDatabaseHandler {
	return &PingDatabaseHandler{service: service}
}

// ServeHTTP accepts GET requests and  checks if database answers to ping command
func (handler PingDatabaseHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()

	if err := handler.service.Ping(ctx); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}
