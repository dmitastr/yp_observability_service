package getmetric

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	srv "github.com/dmitastr/yp_observability_service/internal/domain/service_interface"
	"github.com/dmitastr/yp_observability_service/internal/errs"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/dmitastr/yp_observability_service/internal/presentation/update"
)

type GetMetricHandler struct {
	service srv.ServiceAbstract
}

func NewHandler(s srv.ServiceAbstract) *GetMetricHandler {
	return &GetMetricHandler{service: s}
}

func (handler GetMetricHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var upd update.MetricUpdate

	switch req.Method {
	case http.MethodGet:
		mtype := req.PathValue("mtype")
		name := req.PathValue("name")
		upd, _ = update.New(name, mtype, "1")
	case http.MethodPost:
		if err := json.NewDecoder(req.Body).Decode(&upd); err != nil {
			http.Error(res, err.Error(), http.StatusNotFound)
			return	
		}
		upd.MetricValue = "1"
	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	logger.GetLogger().Infof("receive update=%s", upd)

	if !upd.IsValid() {
		http.Error(res, errs.ErrorWrongPath.Error(), http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()
	metric, err := handler.service.GetMetric(ctx, upd)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}

	switch req.Method {
	case http.MethodGet:
		valString, err := metric.GetValueString()
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(valString))
		return
		
	case http.MethodPost:
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(res).Encode(&metric); err != nil {
			http.Error(res, err.Error(), http.StatusNotFound)
			return	
		}
		return
	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}
