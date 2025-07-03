package updatemetric

import (
	"fmt"
	"net/http"

	"github.com/dmitastr/yp_observability_service/internal/presentation/update"
	srv "github.com/dmitastr/yp_observability_service/internal/domain/service_interface"
)

type MetricHandler struct {
	service srv.ServiceAbstract
}

func NewHandler(s srv.ServiceAbstract) *MetricHandler {
	return &MetricHandler{service: s}
}

func (handler MetricHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	mtype := req.PathValue("mtype")
	name := req.PathValue("name")
	value := req.PathValue("value")
	upd := update.MetricUpdate{MType: mtype, MetricName: name, MetricValue: value}
	fmt.Printf("Receive update: type=%s, name=%s, value=%s\n", mtype, name, value)

	if err := upd.IsValid(); err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}

	err := handler.service.ProcessUpdate(upd)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusOK)

}
