package models

import (
	"fmt"
	"strconv"

	"github.com/dmitastr/yp_observability_service/internal/common"
	"github.com/dmitastr/yp_observability_service/internal/errs"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/dmitastr/yp_observability_service/internal/presentation/update"
	"github.com/jackc/pgx/v5"
)

// Metrics stores information about a single metric. Delta and Value are pointers to distinguish nil value from 0
type Metrics struct {
	ID    string   `json:"id" db:"name"`
	MType string   `json:"type" db:"mtype"`
	Delta *int64   `json:"delta,omitempty" db:"delta"`
	Value *float64 `json:"value,omitempty" db:"value"`
	Hash  string   `json:"-" db:"-"`
}

// FromUpdate converts [update.MetricUpdate] to [Metrics]
func FromUpdate(upd update.MetricUpdate) (m Metrics) {
	m.Delta = upd.Delta
	m.Value = upd.Value
	m.ID = upd.MetricName
	m.MType = upd.MType
	return
}

// UpdateDelta increment Delta value or set it if it's nil
func (m *Metrics) UpdateDelta(value int64) {
	if m.Delta == nil {
		m.Delta = &value
		return
	}
	*m.Delta += value
}

// SetValue updates value field of a metric
func (m *Metrics) SetValue(value *float64) {
	*m.Value = *value
}

// GetValueString select metric value based on its type and converts it to string
func (m *Metrics) GetValueString() (val string, err error) {
	switch m.MType {
	case common.GAUGE:
		if m.Value != nil {
			val = common.FormatFloatTrimZero(*m.Value)
		}
	case common.COUNTER:
		if m.Delta != nil {
			val = strconv.FormatInt(*m.Delta, 10)
		}
	default:
		err = errs.ErrorValueFromEmptyMetric
	}
	return val, err
}

func (m *Metrics) String() string {
	strVal, err := m.GetValueString()
	if err != nil {
		logger.Error(err)
		return ""
	}
	return fmt.Sprintf("name=%s, type=%s, value=%s", m.ID, m.MType, strVal)
}

// ToNamedArgs converts [Metric] to [pgx.NamedArgs] for SQL query
func (m *Metrics) ToNamedArgs() pgx.NamedArgs {
	args := pgx.NamedArgs{
		"name":  m.ID,
		"mtype": m.MType,
		"value": m.Value,
		"delta": m.Delta,
	}
	return args
}
