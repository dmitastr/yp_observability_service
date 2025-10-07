package updatemetricsbatch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	srv "github.com/dmitastr/yp_observability_service/internal/domain/service_interface"
	models "github.com/dmitastr/yp_observability_service/internal/model"
)

type BatchUpdateHandler struct {
	service srv.ServiceAbstract
}

func NewHandler(s srv.ServiceAbstract) *BatchUpdateHandler {
	return &BatchUpdateHandler{service: s}
}

func (handler BatchUpdateHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var metrics []models.Metrics
	if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		http.Error(res, fmt.Errorf("error while decoding request body: %v", err).Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()

	if err := handler.service.BatchUpdate(ctx, metrics); err != nil {
		http.Error(res, fmt.Errorf("error while batch metrics update: %v", err).Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}
