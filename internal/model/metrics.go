package models

import (
	"strconv"

	formattools "github.com/dmitastr/yp_observability_service/internal/format_tools"
	"github.com/dmitastr/yp_observability_service/internal/errs"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func (m *Metrics) DeltaSet(value *int64) {
	if m.Delta == nil {
		m.Delta = value
		return
	}
	*m.Delta += *value
}

func (m *Metrics) GetValueString() (val string, err error) {
	if m.Delta != nil {
		val = strconv.FormatInt(*m.Delta, 10)
	} else if m.Value != nil {
		val = formattools.FormatFloatTrimZero(*m.Value)
	} else {
		err = errs.ErrorValueFromEmptyMetric
	}
	return val, err
}