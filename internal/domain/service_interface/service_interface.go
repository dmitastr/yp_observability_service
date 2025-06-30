package serviceinterface

import "github.com/dmitastr/yp_observability_service/internal/presentation/update"
import "github.com/dmitastr/yp_observability_service/internal/model"

type ServiceAbstract interface {
	ProcessUpdate(update.MetricUpdate) error
	GetMetric(name, mType string) (*models.Metrics, error)
}
