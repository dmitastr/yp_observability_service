package updatemetricsbatch

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dmitastr/yp_observability_service/internal/common"
	"github.com/dmitastr/yp_observability_service/internal/domain/models"
	srv "github.com/dmitastr/yp_observability_service/internal/domain/service"
	"github.com/dmitastr/yp_observability_service/internal/logger"
)

// BatchUpdateHandler handles requests for updating metric value or creating a new one for several metrics
type BatchUpdateHandler struct {
	service srv.IService
}

func NewHandler(s srv.IService) *BatchUpdateHandler {
	return &BatchUpdateHandler{service: s}
}

// ServeHTTP accepts POST requests. It accepts a list of metrics with values in json format. It then updates them
// in db and returns appropriate http code
func (handler BatchUpdateHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var metrics []models.Metrics
	if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		logger.Errorf("error while decoding request body: %v", err)
		http.Error(res, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()

	ctx = context.WithValue(ctx, common.SenderInfo{}, common.ExtractIP(req))

	if err := handler.service.BatchUpdate(ctx, metrics); err != nil {
		logger.Errorf("error while batch metrics update: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}
