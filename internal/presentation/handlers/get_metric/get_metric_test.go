package getmetric

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitastr/yp_observability_service/internal/mocks"
	models "github.com/dmitastr/yp_observability_service/internal/model"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetMetricHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var delta int64 = 10
	value := 99.9
	metric := models.Metrics{ID: "abc", MType: "gauge", Value: &value, Delta: &delta}

	errFunc := func(serviceErrOut bool) (err error) {
		if serviceErrOut {
			err = errors.New("mocked error")
		}
		return
	}

	type pathParam struct {
		Name   string
		Mtype string
	}

	tests := []struct {
		name          string
		method        string
		wantCode      int
		pathParams    pathParam
		wantErr bool
		serviceErrOut bool
	}{
		{
			name:     "Valid request",
			method:   http.MethodGet,
			wantCode: http.StatusOK,
			pathParams: pathParam{
				Name: "abc",
				Mtype: "gauge",
			},
			serviceErrOut: false,
			wantErr: false,
		},
		{
			name:     "POST method",
			method:   http.MethodPost,
			wantCode: http.StatusOK,
			pathParams: pathParam{
				Name: "abc",
				Mtype: "gauge",
			},
			serviceErrOut: false,
			wantErr: false,
		},
		{
			name:     "Bad path - missing param",
			method:   http.MethodGet,
			wantCode: http.StatusNotFound,
			pathParams: pathParam{
				Name: "",
				Mtype: "gauge",
			},
			serviceErrOut: false,
			wantErr: true,
		},
		{
			name:     "Service returned an error",
			method:   http.MethodGet,
			wantCode: http.StatusBadRequest,
			pathParams: pathParam{
				Name: "abc",
				Mtype: "gauge",
			},
			serviceErrOut: true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSrv := mocks.NewMockServiceAbstract(ctrl)
			errValue := errFunc(tt.serviceErrOut)
			mockSrv.EXPECT().GetMetric(context.TODO(), gomock.Any()).Return(&metric, errValue).AnyTimes()
			
			handler := NewHandler(mockSrv)

			req := httptest.NewRequest(tt.method, "http://localhost", nil)
			rr := httptest.NewRecorder()

			req.SetPathValue("name", tt.pathParams.Name)
			req.SetPathValue("mtype", tt.pathParams.Mtype)

			handler.ServeHTTP(rr, req)
			if tt.wantErr {
				assert.Equal(t, tt.wantCode, rr.Code)
				return
			}

			val := rr.Body.String()
			assert.NotEmpty(t, val)
		})
	}
}
