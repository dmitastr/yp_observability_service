package service

import (
	"context"
	"errors"
	"slices"

	"github.com/dmitastr/yp_observability_service/internal/common"
	"github.com/dmitastr/yp_observability_service/internal/domain/audit"
	"github.com/dmitastr/yp_observability_service/internal/domain/audit/data"
	"github.com/dmitastr/yp_observability_service/internal/domain/models"
	"github.com/dmitastr/yp_observability_service/internal/domain/pinger"
	"github.com/dmitastr/yp_observability_service/internal/errs"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/dmitastr/yp_observability_service/internal/presentation/update"
	dbinterface "github.com/dmitastr/yp_observability_service/internal/repository"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	db      dbinterface.Database
	pinger  pinger.Pinger
	auditor audit.IAuditor
}

func NewService(db dbinterface.Database, pinger pinger.Pinger, auditor audit.IAuditor) *Service {
	return &Service{db: db, pinger: pinger, auditor: auditor}
}

func (service Service) ProcessUpdate(ctx context.Context, upd update.MetricUpdate) error {
	logger.Infof("Processing update: %s", upd)
	metricNew := models.FromUpdate(upd)
	metricExist, err := service.db.Get(ctx, metricNew.ID)
	if err != nil {
		return err
	}
	if metricExist != nil && metricExist.Delta != nil {
		metricNew.UpdateDelta(*metricExist.Delta)
	}

	return service.db.Update(ctx, metricNew)
}

func (service Service) BatchUpdate(ctx context.Context, metrics []models.Metrics) error {
	for i, m := range metrics {
		if m.MType != common.COUNTER {
			continue
		}

		mExist, err := service.db.Get(ctx, m.ID)
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Warn("No rows returned")
		} else if err != nil {
			continue
		}

		if mExist != nil && mExist.Delta != nil {
			metrics[i].UpdateDelta(*mExist.Delta)
		}
	}

	if err := service.db.BulkUpdate(ctx, metrics); errors.Is(err, pgx.ErrNoRows) {
		logger.Warn("No rows returned")
	} else if err != nil {
		logger.Errorf("Bulk Update Error: %v", err)
		return err
	}

	var ip string
	ip, ok := ctx.Value(common.SenderInfo{}).(string)
	if !ok {
		ip = "127.0.0.1"
	}

	auditData := data.NewData(metrics, ip)
	if err := service.auditor.Notify(auditData); err != nil {
		logger.Error(err)
	}
	return nil
}

func (service Service) GetMetric(ctx context.Context, upd update.MetricUpdate) (metric *models.Metrics, err error) {
	metric, err = service.db.Get(ctx, upd.MetricName)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	if metric == nil {
		err = errs.ErrorMetricDoesNotExist
	}
	return metric, err
}

func (service Service) GetAll(ctx context.Context) (metricLst []models.DisplayMetric, err error) {
	metricDB, err := service.db.GetAll(ctx)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	for _, m := range metricDB {
		md := models.ModelToDisplay(m)
		metricLst = append(metricLst, md)
	}
	if len(metricLst) == 0 {
		err = errs.ErrorMetricTableEmpty
	}
	slices.SortFunc(metricLst, func(a, b models.DisplayMetric) int {
		if a.Name > b.Name {
			return 1
		}
		return -1
	})
	return metricLst, err
}

func (service Service) Ping(ctx context.Context) error {
	return service.pinger.Ping(ctx, service.db)
}
