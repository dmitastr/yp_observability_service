package metric

import "strconv"

type Metric interface {
	ToString() []any
	UpdateValue(value any)
	GetStringValue() string
}

type GaugeMetric struct {
	ID    string
	MType string
	Value float64
}

func NewGaugeMetric(ID string, Value float64) *GaugeMetric {
	return &GaugeMetric{ID: ID, MType: "gauge", Value: Value}
}

func (m GaugeMetric) ToString() []any {
	pathParams := []any{m.MType, m.ID, m.GetStringValue()}
	return pathParams
}

func (m *GaugeMetric) UpdateValue(value any) {
	newValue := value.(float64)
	m.Value = newValue
}

func (m GaugeMetric) GetStringValue() string {
	return strconv.FormatFloat(m.Value, 'f', 6, 64)
}

type CounterMetric struct {
	ID    string
	MType string
	Value int64
}

func NewCounterMetric(ID string, Value int64) *CounterMetric {
	return &CounterMetric{ID: ID, MType: "counter", Value: Value}
}

func (m CounterMetric) ToString() []any {
	pathParams := []any{m.MType, m.ID, m.GetStringValue()}
	return pathParams
}

func (m CounterMetric) GetStringValue() string {
	return strconv.FormatInt(m.Value, 10)
}

func (m *CounterMetric) UpdateValue(value any) {
	newValue := value.(int64)
	m.Value += newValue
}