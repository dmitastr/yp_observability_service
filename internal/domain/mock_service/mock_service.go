package mockservice

import (
	"errors"

	models "github.com/dmitastr/yp_observability_service/internal/model"
	"github.com/dmitastr/yp_observability_service/internal/presentation/update"
	"github.com/dmitastr/yp_observability_service/internal/domain/entity"
)

var ErrorMock = errors.New("mock error")

type MockService struct {
	WantErr bool
}

func (s MockService) ProcessUpdate(update.MetricUpdate) error {
	if s.WantErr {
		return ErrorMock
	}
	return nil
}

func (s MockService) GetMetric(name, mType string) (*models.Metrics, error) {
	value := 1.0
	delta := int64(1)
	m := models.Metrics{ID: name, MType: mType, Value: &value, Delta: &delta}
	if s.WantErr {
		return &m, ErrorMock
	}
	return &m, nil
}

func (s MockService) GetAll() (lst []entity.DisplayMetric, err error) {
	m := entity.DisplayMetric{Name: "name", Type: "mType", StringValue: "1"}
	lst = append(lst, m)
	if s.WantErr {
		return lst, ErrorMock
	}
	return lst, nil
}

