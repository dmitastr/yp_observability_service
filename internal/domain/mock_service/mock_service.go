package mockservice

import (
	"errors"

	models "github.com/dmitastr/yp_observability_service/internal/model"
	"github.com/dmitastr/yp_observability_service/internal/presentation/update"
)

type MockService struct {
	WantErr bool
}

func (s MockService) ProcessUpdate(update.MetricUpdate) error {
	if s.WantErr {
		return errors.New("mock error")
	}
	return nil
}

func (s MockService) GetMetric(name, mType string) (*models.Metrics, error) {
	value := 1.0
	delta := int64(1)
	m := models.Metrics{ID: name, MType: mType, Value: &value, Delta: &delta}
	if s.WantErr {
		return &m, errors.New("mock error")
	}
	return &m, nil
}
