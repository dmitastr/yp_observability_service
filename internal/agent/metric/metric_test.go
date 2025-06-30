package metric

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestGaugeMetric_ToString(t *testing.T) {
	tests := []struct {
		name string
		m    *GaugeMetric
		want []string
	}{
		{
			name: "valid input",
			m: NewGaugeMetric("abc", 10.0),
			want: []string{"abc", "gauge", strconv.FormatFloat(10, 'f', 6, 64)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ElementsMatch(t, tt.want, tt.m.ToString())
		})
	}
}

func TestCounterMetric_ToString(t *testing.T) {
	tests := []struct {
		name string
		m    *CounterMetric
		want []string
	}{
		{
			name: "valid input",
			m: NewCounterMetric("abc", 10),
			want: []string{"abc", "counter", strconv.FormatInt(10, 10)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ElementsMatch(t, tt.want, tt.m.ToString())
		})
	}
}

func TestGaugeMetric_UpdateValue(t *testing.T) {
	tests := []struct {
		name string
		m    *GaugeMetric
		newValue any
		wantValue float64
		wantErr bool
	}{
		{
			name: "valid input",
			m: NewGaugeMetric("abc", 10),
			newValue: 20.0,
			wantValue: 20.0,
			wantErr: false,
		},
		{
			name: "input is not valid number",
			m: NewGaugeMetric("abc", 10),
			newValue: "10",
			wantValue: 0,
			wantErr: true,
		},
		
		{
			name: "input is not float64",
			m: NewGaugeMetric("abc", 10),
			newValue: 10,
			wantValue: 0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.m.UpdateValue(tt.newValue)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.wantValue, tt.m.Value)
		})
	}
}


func TestCounterMetric_UpdateValue(t *testing.T) {
	tests := []struct {
		name string
		m    *CounterMetric
		newValue any
		wantValue int64
		wantErr bool
	}{
		{
			name: "valid input",
			m: NewCounterMetric("abc", 10),
			newValue: int64(20),
			wantValue: 30,
			wantErr: false,
		},
		{
			name: "input is not valid number",
			m: NewCounterMetric("abc", 10),
			newValue: "20",
			wantValue: 0,
			wantErr: true,
		},
		
		{
			name: "input is not int64",
			m: NewCounterMetric("abc", 10),
			newValue: 10,
			wantValue: 0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.m.UpdateValue(tt.newValue)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.wantValue, tt.m.Value)
		})
	}
}


func TestCounterMetric_GetValue(t *testing.T) {
	tests := []struct {
		name string
		m    *CounterMetric
		key string
		wantValue any
	}{
		{
			name: "valid input",
			m: NewCounterMetric("abc", 10),
			key: "abc",
			wantValue: int64(10),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantValue, tt.m.GetValue())
		})
	}
}


func TestGaugeMetric_GetValue(t *testing.T) {
	tests := []struct {
		name string
		m    *GaugeMetric
		key string
		wantValue any
	}{
		{
			name: "valid input",
			m: NewGaugeMetric("abc", 10),
			key: "abc",
			wantValue: 10.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantValue, tt.m.GetValue())
		})
	}
}
