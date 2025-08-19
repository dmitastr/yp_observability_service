package updatemetric

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	srv "github.com/dmitastr/yp_observability_service/internal/domain/service_interface"
	"github.com/dmitastr/yp_observability_service/internal/errs"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/dmitastr/yp_observability_service/internal/presentation/update"
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
	logger.GetLogger().Infof("Receive update: type=%s, name=%s, value=%s\n", mtype, name, value)

	upd, err := update.New(name, mtype, value)
	if err != nil {
		err = fmt.Errorf("error creating update from request: %v", err)
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if upd.IsEmpty() {
		if err := json.NewDecoder(req.Body).Decode(&upd); err != nil {
			err = fmt.Errorf("error while reading request body: %v", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	} else if !upd.IsValid() {
		http.Error(res, errs.ErrorWrongPath.Error(), http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()

	err = handler.service.ProcessUpdate(ctx, upd)
	if err != nil {
		err = fmt.Errorf("get error while processing update: %v", err)
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusOK)

}
