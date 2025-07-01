package listmetric

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitastr/yp_observability_service/internal/domain/mock_service"
	"github.com/stretchr/testify/assert"
)
func TestListMetricsHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		url           string
		wantCode      int
		wantContent   []string
		serviceErrOut bool
	}{
		{
			name:     "Valid request",
			method:   http.MethodGet,
			url:      "/",
			wantCode: http.StatusOK,
			serviceErrOut: false,
			wantContent: []string{"Name", "Metrics", "Value"},
		},
		{
			name:     "POST method",
			method:   http.MethodPost,
			url:      "/",
			wantCode: http.StatusMethodNotAllowed,
			serviceErrOut: false,
			wantContent: []string{},
		},
		{
			name:     "service returned an error",
			method:   http.MethodGet,
			url:      "/",
			wantCode: http.StatusInternalServerError,
			serviceErrOut: true,
			wantContent: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSrv := mockservice.MockService{WantErr: tt.serviceErrOut}
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
