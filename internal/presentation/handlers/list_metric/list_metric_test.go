package listmetric

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitastr/yp_observability_service/internal/domain/models"
	"github.com/dmitastr/yp_observability_service/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestListMetricsHandler_ServeHTTP(t *testing.T) {
	metrics := []models.DisplayMetric{{Name: "abc", Type: "gauge", StringValue: "10"}}
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
		wantContent   []string
		serviceErrOut bool
	}{
		{
			name:          "Valid request",
			method:        http.MethodGet,
			url:           "/",
			wantCode:      http.StatusOK,
			serviceErrOut: false,
			wantContent:   []string{"Name", "Metrics", "Value"},
		},
		{
			name:          "POST method",
			method:        http.MethodPost,
			url:           "/",
			wantCode:      http.StatusMethodNotAllowed,
			serviceErrOut: false,
			wantContent:   []string{},
		},
		{
			name:          "service returned an error",
			method:        http.MethodGet,
			url:           "/",
			wantCode:      http.StatusInternalServerError,
			serviceErrOut: true,
			wantContent:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSrv := mocks.NewMockServiceAbstract(ctrl)
			errValue := errFunc(tt.serviceErrOut)
			mockSrv.EXPECT().GetAll(gomock.Any()).Return(metrics, errValue).AnyTimes()

			handler := NewHandler(mockSrv)

			req := httptest.NewRequest(tt.method, tt.url, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)
			assert.Equal(t, tt.wantCode, rr.Code)

			if rr.Code == http.StatusOK {
				bodyBytes := rr.Body.String()
				for _, substr := range tt.wantContent {
					assert.Contains(t, bodyBytes, substr)
				}
			}
		})
	}
}
