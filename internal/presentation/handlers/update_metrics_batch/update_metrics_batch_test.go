package updatemetricsbatch

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitastr/yp_observability_service/internal/mocks/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBatchUpdateHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	errFunc := func(serviceErrOut bool) (err error) {
		if serviceErrOut {
			err = errors.New("mocked error")
		}
		return
	}

	tests := []struct {
		name          string
		method        string
		url           string
		wantCode      int
		payload       []byte
		serviceErrOut bool
	}{
		{
			name:     "Valid request",
			method:   http.MethodPost,
			url:      "/updates",
			wantCode: http.StatusOK,
			payload: []byte(`[{
				"id": "abc",
				"type": "gauge",
				"value": 1.99
			},{
				"id": "sdf",
				"type": "counter",
				"delta": 1111
			}]`),
			serviceErrOut: false,
		},
		{
			name:     "bad payload",
			method:   http.MethodPost,
			url:      "/updates",
			wantCode: http.StatusBadRequest,
			payload: []byte(`{
				"id": "abc",
				"type": "gauge",
				"value": 1.99
			}`),
			serviceErrOut: true,
		},
		{
			name:     "app returned an error",
			method:   http.MethodPost,
			url:      "/updates",
			wantCode: http.StatusInternalServerError,
			payload: []byte(`[{
				"id": "abc",
				"type": "gauge",
				"value": 1.99
			}]`),
			serviceErrOut: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBuffer(tt.payload))

			mockSrv := service.NewMockIService(ctrl)
			errValue := errFunc(tt.serviceErrOut)

			mockSrv.EXPECT().BatchUpdate(gomock.Any(), gomock.Any()).Return(errValue).AnyTimes()

			handler := NewHandler(mockSrv)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			assert.Equal(t, tt.wantCode, rr.Code)
		})
	}
}
