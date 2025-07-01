package entity

import (
	models "github.com/dmitastr/yp_observability_service/internal/model"
)

type DisplayMetric struct {
	Name        string
	Type        string
	StringValue string
}

func ModelToDisplay(m models.Metrics) DisplayMetric {
	val, err := m.GetValueString()
	if err != nil {
		val = ""
	}
	return DisplayMetric{Name: m.ID, Type: m.MType, StringValue: val}
}
