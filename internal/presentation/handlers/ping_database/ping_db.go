package pingdatabase

import (
	"context"
	"net/http"
	"time"

	srv "github.com/dmitastr/yp_observability_service/internal/domain/service_interface"
)

type PingDatabaseHandler struct {
	service srv.ServiceAbstract
}

func New(service srv.ServiceAbstract) *PingDatabaseHandler {
	return &PingDatabaseHandler{service: service}	
}

func (handler PingDatabaseHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	
	if err := handler.service.Ping(ctx); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}