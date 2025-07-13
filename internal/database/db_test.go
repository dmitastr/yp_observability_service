package database

import (
	"fmt"
	"testing"

	models "github.com/dmitastr/yp_observability_service/internal/model"
	"github.com/stretchr/testify/assert"
)

type args struct {
	gaugeValue   float64
	counterValue int64
	name         string
	mtype        string
}

func metricsFromArgs(args args) *models.Metrics {
	model := models.Metrics{ID: args.name, MType: args.mtype}
	switch model.MType {
	case "gauge":
		model.Value = &args.gaugeValue
	case "counter":
		model.DeltaSet(&args.counterValue)
	default:
		panic("unrecognized metric type")
	}
	return &model
}

type argsList []args

func TestModelToEntity(t *testing.T) {

	tests := []struct {
		name string
		args args
	}{
		{
			name: "gauge only",
			args: args{gaugeValue: 1.0, counterValue: 99, name: "name", mtype: "gauge"},
		},
		{
			name: "counter only",
			args: args{gaugeValue: 1.0, counterValue: 99, name: "name", mtype: "counter"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := metricsFromArgs(tt.args)
			want := MetricEntity{ID: tt.args.name, MType: tt.args.mtype}
			switch tt.args.mtype {
			case "counter":
				want.CounterValue = tt.args.counterValue
			case "gauge":
				want.GaugeValue = tt.args.gaugeValue
			}
			got := ModelToEntity(*model)
			assert.Equal(t, want, got)
		})
	}
}

func TestStorage_Update(t *testing.T) {
	tests := []struct {
		name    string
		storage *Storage
		args    argsList
	}{
		{
			name:    "one value",
			storage: NewStorage(),
			args:    argsList{{gaugeValue: 1.0, counterValue: 99, name: "name", mtype: "counter"}},
		},

		{
			name:    "several values",
			storage: NewStorage(),
			args: argsList{
				{gaugeValue: 1.0, counterValue: 99, name: "name", mtype: "counter"},
				{gaugeValue: 1.0, counterValue: 99, name: "abc", mtype: "counter"},
				{gaugeValue: 10.0, counterValue: 990, name: "name", mtype: "counter"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := make([]*models.Metrics, 0)
			for _, arg := range tt.args {
				newMetric := metricsFromArgs(arg)
				metrics = append(metrics, newMetric)
				tt.storage.Update(*newMetric)
			}
			lastIdx := len(tt.args) - 1
			lastGot := tt.storage.Get(tt.args[lastIdx].name)
			lastWant := metrics[lastIdx]
			assert.Equal(t, lastWant, lastGot)
		})
	}
}

func TestStorage_GetAll(t *testing.T) {
	metricsGenerate := func(n int) (data []args) {
		for i := range n {
			data = append(data, args{counterValue: int64(i), name: fmt.Sprintf(`name%d`, i), mtype: "counter"})
		}
		return
	}
	tests := []struct {
		name    string
		storage *Storage
		want    []models.Metrics
		args    argsList
	}{
		{
			name:    "one metric",
			storage: NewStorage(),
			args:    metricsGenerate(1),
		},
		{
			name:    "many metrics",
			storage: NewStorage(),
			args:    metricsGenerate(100),
		},
		{
			name:    "zero metrics",
			storage: NewStorage(),
			args:    metricsGenerate(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := make([]models.Metrics, 0)
			for _, inp := range tt.args {
				m := metricsFromArgs(inp)
				want = append(want, *m)
				tt.storage.Update(*m)
			}
			got := tt.storage.GetAll()
			assert.ElementsMatch(t, want, got)
		})
	}
}

func TestStorage_Get(t *testing.T) {

	tests := []struct {
		name    string
		storage *Storage
		key     string
		args    args
		wantNil bool
	}{
		{
			name:    "counter metrics exist",
			storage: NewStorage(),
			key:     "name",
			args:    args{counterValue: 10, name: "name", mtype: "counter"},
		},
		{
			name:    "gauge metrics exist",
			key:     "name",
			storage: NewStorage(),
			args:    args{gaugeValue: 10.101, name: "name", mtype: "gauge"},
		},
		{
			name:    "metric doesn't exist",
			key:     "newname",
			storage: NewStorage(),
			args:    args{gaugeValue: 10.101, name: "name", mtype: "gauge"},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric := metricsFromArgs(tt.args)
			tt.storage.Update(*metric)
			got := tt.storage.Get(tt.key)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			assert.Equal(t, metric, got)
		})
	}
}
