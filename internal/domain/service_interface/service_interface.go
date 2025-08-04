package serviceinterface

import (
	"context"

	"github.com/dmitastr/yp_observability_service/internal/domain/entity"
	"github.com/dmitastr/yp_observability_service/internal/model"
	"github.com/dmitastr/yp_observability_service/internal/presentation/update"
)

type ServiceAbstract interface {
	ProcessUpdate(context.Context, update.MetricUpdate) error
	BatchUpdate(context.Context, []models.Metrics) error
	GetMetric(context.Context, update.MetricUpdate) (*models.Metrics, error)
	GetAll(context.Context) ([]entity.DisplayMetric, error)
	Ping(context.Context) error
}
