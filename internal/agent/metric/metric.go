package metric

import (
	"fmt"
	"strconv"

	formattools "github.com/dmitastr/yp_observability_service/internal/format_tools"
)

type Metric interface {
	ToString() [3]string
	UpdateValue(value any) error
	GetStringValue() string
	GetValue() any
}

type GaugeMetric struct {
	ID    string  `json:"id"`
	MType string  `json:"type"`
	Value float64 `json:"value"`
}

func NewGaugeMetric(ID string, Value float64) *GaugeMetric {
	return &GaugeMetric{ID: ID, MType: "gauge", Value: Value}
}

func (m GaugeMetric) ToString() [3]string {
	pathParams := [3]string{m.MType, m.ID, m.GetStringValue()}
	return pathParams
}

func (m *GaugeMetric) UpdateValue(value any) error {
	if newValue, ok := value.(float64); ok {
		m.Value = newValue
		return nil
	}
	return fmt.Errorf("wrong value: expected float64, got %v", value)
}

func (m GaugeMetric) GetStringValue() string {
	return formattools.FormatFloatTrimZero(m.Value)
}

func (m GaugeMetric) GetValue() any {
	return m.Value
}

type CounterMetric struct {
	ID    string `json:"id"`
	MType string `json:"type"`
	Value int64  `json:"delta"`
}

func NewCounterMetric(ID string, Value int64) *CounterMetric {
	return &CounterMetric{ID: ID, MType: "counter", Value: Value}
}

func (m CounterMetric) ToString() [3]string {
	pathParams := [3]string{m.MType, m.ID, m.GetStringValue()}
	return pathParams
}

func (m CounterMetric) GetStringValue() string {
	return strconv.FormatInt(m.Value, 10)
}

func (m *CounterMetric) UpdateValue(value any) error {
	if newValue, ok := value.(int64); ok {
		m.Value += newValue
		return nil
	}
	return fmt.Errorf("wrong value: expected int64, got %v", value)
}

func (m CounterMetric) GetValue() any {
	return m.Value
}
