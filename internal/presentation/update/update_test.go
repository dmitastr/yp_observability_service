package update

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	var (
		delta int64 = 10
		value float64 = 99.9
	)
	type args struct {
		name  string
		mtype string
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    MetricUpdate
		wantErr bool
	}{
		{
			name: "update with counter",
			args: args{name: "abc", mtype: "counter", value: "10"},
			want: MetricUpdate{MType: "counter", MetricName: "abc", MetricValue: "10", Delta: &delta},
			wantErr: false,
		},
		{
			name: "update with gauge",
			args: args{name: "abc", mtype: "gauge", value: "99.9"},
			want: MetricUpdate{MType: "gauge", MetricName: "abc", MetricValue: "99.9", Value: &value},
			wantErr: false,
		},
		{
			name: "gauge wrong value",
			args: args{name: "abc", mtype: "gauge", value: "abc"},
			wantErr: true,
		},
		{
			name: "counter wrong value",
			args: args{name: "abc", mtype: "counter", value: "abc"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := New(tt.args.name, tt.args.mtype, tt.args.value)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, m)
		})
	}
}

func TestMetricUpdate_IsValid(t *testing.T) {
	tests := []struct {
		name string
		upd  MetricUpdate
		want bool
	}{
		{
			name: "valid update",
			upd:  MetricUpdate{MType: "abc", MetricName: "abc", MetricValue: "abc"},
			want: true,
		},
		{
			name: "missing type",
			upd:  MetricUpdate{MType: "", MetricName: "abc", MetricValue: "abc"},
			want: false,
		},
		{
			name: "missing name",
			upd:  MetricUpdate{MType: "abc", MetricName: "", MetricValue: "abc"},
			want: false,
		},
		{
			name: "missing value",
			upd:  MetricUpdate{MType: "abc", MetricName: "abc", MetricValue: ""},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.upd.IsValid())
		})
	}
}

func TestMetricUpdate_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		upd  MetricUpdate
		want bool
	}{
		{
			name: "full update",
			upd:  MetricUpdate{MType: "abc", MetricName: "abc", MetricValue: "abc"},
			want: false,
		},
		{
			name: "missing type",
			upd:  MetricUpdate{MType: "", MetricName: "abc", MetricValue: "abc"},
			want: false,
		},
		{
			name: "missing name",
			upd:  MetricUpdate{MType: "abc", MetricName: "", MetricValue: "abc"},
			want: false,
		},
		{
			name: "missing value",
			upd:  MetricUpdate{MType: "abc", MetricName: "abc", MetricValue: ""},
			want: false,
		},
		{
			name: "missing all",
			upd:  MetricUpdate{MType: "", MetricName: "", MetricValue: ""},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.upd.IsEmpty())
		})
	}
}
