package update

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricUpdate_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		upd     MetricUpdate
		wantErr bool
	}{
		{
			name: "valid update",
			upd: MetricUpdate{MType: "abc", MetricName: "abc", MetricValue: "abc"},
			wantErr: false,
		},
		{
			name: "missing type",
			upd: MetricUpdate{MType: "", MetricName: "abc", MetricValue: "abc"},
			wantErr: true,
		},
		{
			name: "missing name",
			upd: MetricUpdate{MType: "abc", MetricName: "", MetricValue: "abc"},
			wantErr: true,
		},
		{
			name: "missing value",
			upd: MetricUpdate{MType: "abc", MetricName: "abc", MetricValue: ""},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.Error(t, tt.upd.IsValid())
				return
			}
			assert.NoError(t, tt.upd.IsValid())
		})
	}
}
