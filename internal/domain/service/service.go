package service

import (
	"github.com/dmitastr/yp_observability_service/internal/domain/entity"
	"github.com/dmitastr/yp_observability_service/internal/errs"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	models "github.com/dmitastr/yp_observability_service/internal/model"
	"github.com/dmitastr/yp_observability_service/internal/presentation/update"
	"github.com/dmitastr/yp_observability_service/internal/repository"
)


type Service struct {
	db repository.Database
}

func NewService(db repository.Database) *Service {
	return &Service{db: db}
}

func (service Service) ProcessUpdate(upd update.MetricUpdate) error {
	logger.GetLogger().Infof("Processing update: %s", upd)
	metric := models.FromUpdate(upd)
	service.db.Update(metric)
	return nil	
}

func (service Service) GetMetric(name, mType string) (metric *models.Metrics, err error) {
	metric = service.db.Get(name)
	if metric == nil {
		err = errs.ErrorMetricDoesNotExist
	}
	return metric, err
}

func (service Service) GetAll() (metricLst []entity.DisplayMetric, err error) {
	metricDB := service.db.GetAll()
	for _, m := range metricDB {
		md := entity.ModelToDisplay(m)
		metricLst = append(metricLst, md)
	}
	if len(metricLst) == 0 {
		err = errs.ErrorMetricTableEmpty
	}
	return metricLst, err
}