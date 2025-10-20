package repository

import (
	"context"

	"github.com/dmitastr/yp_observability_service/internal/domain/model"
)

type Database interface {
	Update(context.Context, models.Metrics) error
	BulkUpdate(context.Context, []models.Metrics) error
	GetAll(context.Context) ([]models.Metrics, error)
	Get(context.Context, string) (*models.Metrics, error)
	Close() error
	Init() error
	Ping(context.Context) error
}
