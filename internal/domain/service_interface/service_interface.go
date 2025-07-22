package serviceinterface

import (
	"github.com/dmitastr/yp_observability_service/internal/presentation/update"
	"github.com/dmitastr/yp_observability_service/internal/model"
	"github.com/dmitastr/yp_observability_service/internal/domain/entity"
)

type ServiceAbstract interface {
	ProcessUpdate(update.MetricUpdate) error
	GetMetric(update.MetricUpdate) (*models.Metrics, error)
	GetAll() ([]entity.DisplayMetric, error)
}
