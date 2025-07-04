package listmetric

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	srv "github.com/dmitastr/yp_observability_service/internal/domain/service_interface"
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

	metrics, err := handler.service.GetAll()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Receive metrics: %v\n", metrics)
	
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
		http.Error(res, "Template parsing error", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	res.WriteHeader(http.StatusOK)
	err = t.Execute(res, metrics)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}
