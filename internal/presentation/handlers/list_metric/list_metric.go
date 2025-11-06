package listmetric

import (
	"context"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	srv "github.com/dmitastr/yp_observability_service/internal/domain/service"
	"github.com/dmitastr/yp_observability_service/internal/logger"
)

// ListMetricsHandler handles requests for getting a list of all metrics
type ListMetricsHandler struct {
	service srv.IService
}

func NewHandler(s srv.IService) *ListMetricsHandler {
	return &ListMetricsHandler{service: s}
}

// ServeHTTP accept GET requests, fetching a list of all metrics from db and
// rendering them as html table
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
		logger.Errorf("error while getting metrics: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	logger.Infof("Receive %d metrics from db", len(metrics))

	wd, err := os.Getwd()
	if err != nil {
		logger.Errorf("error while getting working directory: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	parentPath := "./"
	if strings.Contains(wd, "handlers") {
		parentPath = "../../../../"
	}
	templatePath := filepath.Join(parentPath, "web", "templates", "index.html")
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		logger.Errorf("Template parsing error: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	err = t.Execute(res, metrics)
	if err != nil {
		logger.Errorf("Template execution error: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
