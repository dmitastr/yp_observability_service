package service

import (
	"strconv"

	"github.com/dmitastr/yp_observability_service/internal/errs"
	models "github.com/dmitastr/yp_observability_service/internal/model"
	"github.com/dmitastr/yp_observability_service/internal/repository"
	"github.com/dmitastr/yp_observability_service/internal/handlers/update"
)

const (
	GAUGE   = "gauge"
	COUNTER = "counter"
)

type ServiceAbstract interface {
	ProcessUpdate(update.MetricUpdate) error
}

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