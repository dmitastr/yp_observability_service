package getmetric

import (
	"net/http"

	srv "github.com/dmitastr/yp_observability_service/internal/domain/service_interface"
	"github.com/dmitastr/yp_observability_service/internal/errs"
)

type GetMetricHandler struct {
	service srv.ServiceAbstract
}

func NewHandler(s srv.ServiceAbstract) *GetMetricHandler {
	return &GetMetricHandler{service: s}
}

func (handler GetMetricHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	mtype := req.PathValue("mtype")
	name := req.PathValue("name")

	if mtype == "" || name == "" {
		http.Error(res, errs.ErrorWrongPath.Error(), http.StatusNotFound)
		return
	}

	metric, err := handler.service.GetMetric(name, mtype)
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
