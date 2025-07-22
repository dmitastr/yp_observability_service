package update

import (
	"fmt"
	"strconv"

	"github.com/dmitastr/yp_observability_service/internal/errs"
)

const (
	GAUGE   = "gauge"
	COUNTER = "counter"
)

type MetricUpdate struct {
	MType       string `json:"type"`
	MetricName  string `json:"id"`
	MetricValue string
	Value       *float64 `json:"value,omitempty"`
	Delta       *int64   `json:"delta,omitempty"`
}

func New(name, mtype, valueStr string) (metric MetricUpdate, err error) {
	metric.MetricName = name
	metric.MType = mtype
	metric.MetricValue = valueStr

	if !metric.IsValid() {
		return
	}

	switch mtype {
	case GAUGE:
		meticValue, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return metric, err
		}
		metric.Value = &meticValue
	case COUNTER:
		meticValue, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return metric, err
		}
		metric.Delta = &meticValue
	default:
		return metric, errs.ErrorWrongUpdateType
	}

	return
}

func (upd MetricUpdate) IsValid() bool {
	return upd.MetricName != "" && upd.MType != "" && upd.MetricValue != ""
}

func (upd MetricUpdate) IsEmpty() bool {
	return upd.MetricName == "" && upd.MType == "" && upd.MetricValue == ""
}

func (upd MetricUpdate) String() string {
	return fmt.Sprintf(`update: name=%s, mtype=%s, value=%s`, upd.MetricName, upd.MType, upd.MetricValue)
}
