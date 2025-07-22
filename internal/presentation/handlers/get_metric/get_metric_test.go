package getmetric

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitastr/yp_observability_service/internal/domain/mock_service"
	"github.com/stretchr/testify/assert"
)

func TestGetMetricHandler_ServeHTTP(t *testing.T) {
	type pathParam struct {
		key   string
		value string
	}

	tests := []struct {
		name          string
		method        string
		wantCode      int
		pathParams    []pathParam
		wantErr bool
		serviceErrOut bool
	}{
		{
			name:     "Valid request",
			method:   http.MethodGet,
			wantCode: http.StatusOK,
			pathParams: []pathParam{
				{key: "name", value: "abc"},
				{key: "mtype", value: "gauge"},
			},
			serviceErrOut: false,
			wantErr: false,
		},
		{
			name:     "POST method",
			method:   http.MethodPost,
			wantCode: http.StatusOK,
			pathParams: []pathParam{
				{key: "name", value: "abc"},
				{key: "mtype", value: "gauge"},
			},
			serviceErrOut: false,
			wantErr: false,
		},
		{
			name:     "Bad path - missing param",
			method:   http.MethodGet,
			wantCode: http.StatusNotFound,
			pathParams: []pathParam{
				{key: "name", value: ""},
				{key: "mtype", value: "gauge"},
			},
			serviceErrOut: false,
			wantErr: true,
		},
		{
			name:     "Service returned an error",
			method:   http.MethodGet,
			wantCode: http.StatusBadRequest,
			pathParams: []pathParam{
				{key: "name", value: "abc"},
				{key: "mtype", value: "gauge"},
			},
			serviceErrOut: true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSrv := mockservice.MockService{WantErr: tt.serviceErrOut}
			handler := NewHandler(mockSrv)

			req := httptest.NewRequest(tt.method, "http://localhost", nil)
			rr := httptest.NewRecorder()
			for _, pp := range tt.pathParams {
				req.SetPathValue(pp.key, pp.value)
			}
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
