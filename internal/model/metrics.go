package models

import (
	"fmt"
	"strconv"

	"github.com/dmitastr/yp_observability_service/internal/errs"
	formattools "github.com/dmitastr/yp_observability_service/internal/format_tools"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/dmitastr/yp_observability_service/internal/presentation/update"
	"github.com/jackc/pgx/v5"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
	Any     = "any"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id" db:"name"`
	MType string   `json:"type" db:"mtype"`
	Delta *int64   `json:"delta,omitempty" db:"delta"`
	Value *float64 `json:"value,omitempty" db:"value"`
	Hash  string   `json:"-" db:"-"`
}

func FromUpdate(upd update.MetricUpdate) (m Metrics) {
	m.Delta = upd.Delta
	m.Value = upd.Value
	m.ID = upd.MetricName
	m.MType = upd.MType
	return
}

func (m *Metrics) DeltaSet(value *int64) {
	if m.Delta == nil {
		m.Delta = value
		return
	}
	*m.Delta += *value
}

func (m *Metrics) GetValueString(mtype string) (val string, err error) {
	switch mtype {
	case Counter:
		if m.Delta != nil {
			val = strconv.FormatInt(*m.Delta, 10)
		}
	case Gauge:
		if m.Value != nil {
			val = formattools.FormatFloatTrimZero(*m.Value)
		}
	case Any:
		if val, err = m.GetValueString(Counter); val == "" {
			return m.GetValueString(Gauge)
		}
	default:
		err = errs.ErrorValueFromEmptyMetric
	}
	return val, err
}

func (m *Metrics) String() string {
	strVal, err := m.GetValueString(Any)
	if err != nil {
		logger.GetLogger().Error(err)
		return ""
	}
	return fmt.Sprintf("name=%s, type=%s, value=%s", m.ID, m.MType, strVal)
}

func (m *Metrics) ToNamedArgs() pgx.NamedArgs {
	args := pgx.NamedArgs{
		"name":  m.ID,
		"mtype": m.MType,
		"value": m.Value,
		"delta": m.Delta,
	}
	return args
}
