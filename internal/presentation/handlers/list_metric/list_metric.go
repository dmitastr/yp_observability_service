package listmetric

import (
	"context"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	srv "github.com/dmitastr/yp_observability_service/internal/domain/service_interface"
	"github.com/dmitastr/yp_observability_service/internal/logger"
)

type ListMetricsHandler struct {
	service srv.ServiceAbstract
}

func NewHandler(s srv.ServiceAbstract) *ListMetricsHandler {
	return &ListMetricsHandler{service: s}
}

func (handler ListMetricsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	res.Header().Set("Content-Type", "text/html")

	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()

	metrics, err := handler.service.GetAll(ctx)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.GetLogger().Infof("Receive %d metrics from db", len(metrics))

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	parentPath := "./"
	if strings.Contains(wd, "handlers") {
		parentPath = "../../../../"
	}
	templatePath := filepath.Join(parentPath, "web", "templates", "index.html")
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		logger.GetLogger().Error(err)
		http.Error(res, "Template parsing error", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	err = t.Execute(res, metrics)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}
