package service

import (
	"errors"
	"strconv"

	"github.com/dmitastr/yp_observability_service/internal/errs"
	models "github.com/dmitastr/yp_observability_service/internal/model"
	"github.com/dmitastr/yp_observability_service/internal/presentation/update"
	"github.com/dmitastr/yp_observability_service/internal/domain/entity"
	"github.com/dmitastr/yp_observability_service/internal/repository"
)

const (
	GAUGE   = "gauge"
	COUNTER = "counter"
)


type Service struct {
	db repository.Database
}

func NewService(db repository.Database) *Service {
	return &Service{db: db}
}

func (service Service) ProcessUpdate(upd update.MetricUpdate) error {
	metric := models.Metrics{ID: upd.MetricName, MType: upd.MType}

	switch upd.MType {
	case GAUGE:
		meticValue, err := strconv.ParseFloat(upd.MetricValue, 64)
		if err != nil {
			return err
		}
		metric.Value = &meticValue
	case COUNTER:
		meticValue, err := strconv.ParseInt(upd.MetricValue, 10, 64)
		if err != nil {
			return err
		}
		metric.Delta = &meticValue
	default:
		return errs.ErrorWrongUpdateType
	}
	service.db.Update(metric)
	return nil	
}


func (service Service) GetMetric(name, mType string) (metric *models.Metrics, err error) {
	metric = service.db.Get(name)
	if metric == nil {
		err = errors.New("metric not found")
	}
	return metric, err
}

func (service Service) GetAll() (metricLst []entity.DisplayMetric, err error) {
	metricDb := service.db.GetAll()
	for _, m := range metricDb {
		md := entity.ModelToDisplay(m)
		metricLst = append(metricLst, md)
	}
	if len(metricLst) == 0 {
		err = errors.New("no metrics added yet")
	}
	return metricLst, err
}