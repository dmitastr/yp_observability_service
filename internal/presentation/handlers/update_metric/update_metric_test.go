package updatemetric

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitastr/yp_observability_service/internal/mocks/service"
	_ "github.com/dmitastr/yp_observability_service/internal/presentation/update"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMetricHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	errFunc := func(serviceErrOut bool) (err error) {
		if serviceErrOut {
			err = errors.New("mocked error")
		}
		return
	}

	type pathParam struct {
		key   string
		value string
	}

	tests := []struct {
		name          string
		method        string
		url           string
		wantCode      int
		pathParams    []pathParam
		serviceErrOut bool
	}{
		{
			name:     "Valid request",
			method:   http.MethodPost,
			url:      "/update/abc/10",
			wantCode: http.StatusOK,
			pathParams: []pathParam{
				{key: "name", value: "abc"},
				{key: "mtype", value: "gauge"},
				{key: "value", value: "10"},
			},
			serviceErrOut: false,
		},
		{
			name:     "Get method",
			method:   http.MethodGet,
			url:      "/update/abc/gauge/10",
			wantCode: http.StatusMethodNotAllowed,
			pathParams: []pathParam{
				{key: "name", value: "abc"},
				{key: "mtype", value: "gauge"},
				{key: "value", value: "10"},
			},
			serviceErrOut: false,
		},
		{
			name:     "Bad path - missing param",
			method:   http.MethodPost,
			url:      "/update/gauge/10",
			wantCode: http.StatusNotFound,
			pathParams: []pathParam{
				{key: "name", value: ""},
				{key: "mtype", value: "gauge"},
				{key: "value", value: "10"},
			},
			serviceErrOut: false,
		},
		{
			name:     "Service returned an error",
			method:   http.MethodPost,
			url:      "/update/abc/10",
			wantCode: http.StatusBadRequest,
			pathParams: []pathParam{
				{key: "name", value: "abc"},
				{key: "mtype", value: "gauge"},
				{key: "value", value: "10"},
			},
			serviceErrOut: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)

			mockSrv := service.NewMockServiceAbstract(ctrl)
			errValue := errFunc(tt.serviceErrOut)

			mockSrv.EXPECT().ProcessUpdate(gomock.Any(), gomock.Any()).Return(errValue).AnyTimes()

			handler := NewHandler(mockSrv)

			rr := httptest.NewRecorder()
			for _, pp := range tt.pathParams {
				req.SetPathValue(pp.key, pp.value)
			}
			handler.ServeHTTP(rr, req)
			assert.Equal(t, tt.wantCode, rr.Code)
		})
	}
}
