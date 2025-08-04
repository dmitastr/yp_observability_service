package service

import (
	"context"
	"slices"

	"github.com/dmitastr/yp_observability_service/internal/domain/entity"
	"github.com/dmitastr/yp_observability_service/internal/errs"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	models "github.com/dmitastr/yp_observability_service/internal/model"
	"github.com/dmitastr/yp_observability_service/internal/presentation/update"
	"github.com/dmitastr/yp_observability_service/internal/repository"
)

type Service struct {
	db   repository.Database
}

func NewService(db repository.Database) *Service {
	return &Service{db: db}
}

func (service Service) ProcessUpdate(ctx context.Context, upd update.MetricUpdate) error {
	logger.GetLogger().Infof("Processing update: %s", upd)
	metric := models.FromUpdate(upd)
	err := service.db.Update(ctx, metric)
	return err
}

func (service Service) BatchUpdate(ctx context.Context, metrics []models.Metrics) error {
	return service.db.BulkUpdate(ctx, metrics)
}

func (service Service) GetMetric(ctx context.Context, upd update.MetricUpdate) (metric *models.Metrics, err error) {
	metric, err = service.db.Get(ctx, upd.MetricName)
	if err != nil {
		return nil, err
	}

	if metric == nil {
		err = errs.ErrorMetricDoesNotExist
	}
	return metric, err
}

func (service Service) GetAll(ctx context.Context) (metricLst []entity.DisplayMetric, err error) {
	metricDB, err := service.db.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, m := range metricDB {
		md := entity.ModelToDisplay(m)
		metricLst = append(metricLst, md)
	}
	if len(metricLst) == 0 {
		err = errs.ErrorMetricTableEmpty
	}
	slices.SortFunc(metricLst, func(a, b entity.DisplayMetric) int {
		if a.Name > b.Name {
			return 1
		}
		return -1
	})
	return metricLst, err
}

func (service Service) Ping(ctx context.Context) error {
	return service.db.Ping(ctx)
}
