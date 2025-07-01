package metric

import (
	"fmt"
	"strconv"
)

type Metric interface {
	ToString() [3]string
	UpdateValue(value any) error
	GetStringValue() string
	GetValue() any
}

type GaugeMetric struct {
	ID    string
	MType string
	Value float64
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
	return strconv.FormatFloat(m.Value, 'f', 3, 64)
}

func (m GaugeMetric) GetValue() any {
	return m.Value
}

type CounterMetric struct {
	ID    string
	MType string
	Value int64
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