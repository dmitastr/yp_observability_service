package getmetric

import (
	"encoding/json"
	"net/http"

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

	metric, err := handler.service.GetMetric(upd)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}
	valString, err := metric.GetValueString()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(valString))

}
