package update

import "github.com/dmitastr/yp_observability_service/internal/errs"

type MetricUpdate struct {
	MType       string
	MetricName  string
	MetricValue string
}

func (upd MetricUpdate) IsValid() error {
	if upd.MetricName == "" || upd.MType == "" || upd.MetricValue == "" {
		return errs.ErrorWrongPath
	}
	return nil
}
