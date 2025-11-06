package example

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dmitastr/yp_observability_service/internal/domain/models"
	"github.com/dmitastr/yp_observability_service/internal/domain/service"
	service2 "github.com/dmitastr/yp_observability_service/internal/mocks/service"
	getmetric "github.com/dmitastr/yp_observability_service/internal/presentation/handlers/get_metric"
	pingdatabase "github.com/dmitastr/yp_observability_service/internal/presentation/handlers/ping_database"
	updatemetric "github.com/dmitastr/yp_observability_service/internal/presentation/handlers/update_metric"
	updatemetricsbatch "github.com/dmitastr/yp_observability_service/internal/presentation/handlers/update_metrics_batch"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
)

func serviceSetup() service.IService {
	value := float64(10)
	var err error
	metric := models.Metrics{
		ID:    "abc",
		MType: "gauge",
		Value: &value,
	}

	ctrl := gomock.NewController(&testing.T{})
	observabilityService := service2.NewMockServiceAbstract(ctrl)
	observabilityService.EXPECT().GetMetric(gomock.Any(), gomock.Any()).Return(&metric, err).AnyTimes()
	observabilityService.EXPECT().ProcessUpdate(gomock.Any(), gomock.Any()).Return(err).AnyTimes()
	observabilityService.EXPECT().BatchUpdate(gomock.Any(), gomock.Any()).Return(err).AnyTimes()
	observabilityService.EXPECT().Ping(gomock.Any()).Return(err).AnyTimes()

	return observabilityService
}

func routerSetup(service service.IService) http.Handler {
	metricHandler := updatemetric.NewHandler(service)
	metricBatchHandler := updatemetricsbatch.NewHandler(service)
	getMetricHandler := getmetric.NewHandler(service)
	pingHandler := pingdatabase.New(service)
	router := chi.NewRouter()
	// setting routes

	router.Route(`/update`, func(r chi.Router) {
		r.Post(`/`, metricHandler.ServeHTTP)
		r.Post(`/{mtype}/{name}/{value}`, metricHandler.ServeHTTP)
	})

	router.Post(`/updates`, metricBatchHandler.ServeHTTP)
	router.Route(`/value`, func(r chi.Router) {
		r.Post(`/`, getMetricHandler.ServeHTTP)
		r.Get(`/{mtype}/{name}`, getMetricHandler.ServeHTTP)
	})

	router.Get(`/ping`, pingHandler.ServeHTTP)
	return router

}

func Example() {
	// prepare mock service for handlers to inject
	observabilityService := serviceSetup()
	// create router with all the handlers and inject service there
	router := routerSetup(observabilityService)
	// create test server to test endpoints
	server := httptest.NewServer(router)
	defer server.Close()

	metricGet := `{"id": "abc", "type": "gauge"}`
	metricPost := `{"id":"abc","type": "gauge","value": 10}`
	metricPostBatch := `[{"id":"gauge_0","type":"gauge","value":44.92},{"id":"counter_1","type":"counter","delta":43}]`

	client := server.Client()
	var err error

	// ping server status
	respPing, err := client.Get(server.URL + `/ping`)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer respPing.Body.Close()
	fmt.Println(respPing.Status)

	// update single metric
	respUpdate, err := client.Post(server.URL+"/update", "application/json", strings.NewReader(metricPost))
	if err != nil {
		return
	}
	defer respUpdate.Body.Close()
	fmt.Println(respUpdate.Status)

	// update several metrics
	respUpdates, err := client.Post(server.URL+"/updates", "application/json", strings.NewReader(metricPostBatch))
	if err != nil {
		return
	}
	defer respUpdates.Body.Close()
	fmt.Println(respUpdates.Status)

	// get single metric
	respGet, err := client.Post(server.URL+"/value", "application/json", strings.NewReader(metricGet))
	if err != nil {
		return
	}
	defer respGet.Body.Close()
	body, _ := io.ReadAll(respGet.Body)
	fmt.Println(string(body))

	// Output:
	// 200 OK
	// 200 OK
	// 200 OK
	// {"id":"abc","type":"gauge","value":10}
}
