package service

import (
	"context"

	"github.com/dmitastr/yp_observability_service/internal/domain/model"
	"github.com/dmitastr/yp_observability_service/internal/presentation/update"
)

type IService interface {
	ProcessUpdate(context.Context, update.MetricUpdate) error
	BatchUpdate(context.Context, []models.Metrics) error
	GetMetric(context.Context, update.MetricUpdate) (*models.Metrics, error)
	GetAll(context.Context) ([]models.DisplayMetric, error)
	Ping(context.Context) error
}
