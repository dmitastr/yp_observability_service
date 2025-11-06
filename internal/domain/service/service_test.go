package service

import (
	"errors"
	"testing"

	mockaudit "github.com/dmitastr/yp_observability_service/internal/mocks/audit"
	mockpinger "github.com/dmitastr/yp_observability_service/internal/mocks/pinger"
	"github.com/dmitastr/yp_observability_service/internal/mocks/storage"
	"github.com/dmitastr/yp_observability_service/internal/presentation/update"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestService_ProcessUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	auditor := mockaudit.NewMockIAuditor(ctrl)
	pinger := mockpinger.NewMockPinger(ctrl)
	db := storage.NewMockDatabase(ctrl)

	type args struct {
		mName, mType string
		mValue       float64
		mDelta       int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "valid input",
			args:    args{mType: "gauge", mName: "abc", mValue: 10},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upd := update.MetricUpdate{MetricName: tt.args.mName, MType: tt.args.mType, Value: &tt.args.mValue}

			var dbErr error
			if tt.wantErr {
				dbErr = errors.New("error")
			}
			db.EXPECT().Get(t.Context(), upd.MetricName).Return(nil, dbErr).AnyTimes()
			db.EXPECT().Update(t.Context(), gomock.Any()).Return(dbErr).AnyTimes()

			observabilityService := NewService(db, pinger, auditor)

			err := observabilityService.ProcessUpdate(t.Context(), upd)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

		})
	}
}
