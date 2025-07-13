package database

import (
	"sync"

	"github.com/dmitastr/yp_observability_service/internal/logger"
	models "github.com/dmitastr/yp_observability_service/internal/model"
)

type MetricEntity struct {
	ID           string
	MType        string
	GaugeValue   float64
	CounterValue int64
}

func ModelToEntity(metric models.Metrics) MetricEntity {
	entity := MetricEntity{
		ID:    metric.ID,
		MType: metric.MType,
	}
	if metric.Value != nil {
		entity.GaugeValue = *metric.Value
	}
	if metric.Delta != nil {
		entity.CounterValue = *metric.Delta
	}

	return entity
}

type Storage struct {
	sync.Mutex
	Metrics map[string]models.Metrics
}

func NewStorage() *Storage {
	s := Storage{Metrics: make(map[string]models.Metrics)}
	return &s
}

func (storage *Storage) Update(newMetric models.Metrics) {
	logger.GetLogger().Infof("Get new data: %s", &newMetric)
	storage.Lock()
	defer storage.Unlock()
	if metric, ok := storage.Metrics[newMetric.ID]; ok {
		if metric.Delta != nil {
			newMetric.DeltaSet(metric.Delta)
		}
	}
	storage.Metrics[newMetric.ID] = newMetric
}

func (storage *Storage) GetAll() (lst []models.Metrics) {
	for _, metric := range storage.Metrics {
		lst = append(lst, metric)
	}

	return
}

func (storage *Storage) Get(key string) *models.Metrics {
	if metric, ok := storage.Metrics[key]; ok {
		logger.GetLogger().Infof("Found metric: %s", metric)
		return &metric
	}
	return nil
}
