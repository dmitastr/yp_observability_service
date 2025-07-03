package database

import (
	"fmt"
	"slices"

	models "github.com/dmitastr/yp_observability_service/internal/model"
)


type MetricEntity struct {
	ID string
	MType string
	GaugeValue float64
	CounterValue int64
}


func ModelToEntity(metric models.Metrics) MetricEntity {
	entity := MetricEntity{
		ID: metric.ID, 
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
	Metrics map[string]models.Metrics
}

func NewStorage() *Storage {
	s := Storage{Metrics: make(map[string]models.Metrics)}
	return &s
}

func (storage *Storage) Update(newMetric models.Metrics) {
	if metric, ok := storage.Metrics[newMetric.ID]; ok {
		if metric.Delta != nil {
			newMetric.DeltaSet(metric.Delta)
		}
	}
	storage.Metrics[newMetric.ID] = newMetric
	fmt.Println(storage.toList())
}




func (storage Storage) toList() (lst []MetricEntity) {
	for _, metric := range storage.Metrics {
		m := ModelToEntity(metric)
		lst = append(lst, m)
	}
	return 
}


func (storage Storage) toList1() (lst []models.Metrics) {
	for _, metric := range storage.Metrics {
		lst = append(lst, metric)
	}
	slices.SortFunc(lst, func(a, b models.Metrics) int {
		if a.ID > b.ID {
			return 1
		}
		return -1
	})
	return 
}


func (storage *Storage) GetAll() []models.Metrics {
	return storage.toList1()
}



func (storage *Storage) Get(key string) *models.Metrics {
	if metric, ok := storage.Metrics[key]; ok {
		val, _ := metric.GetValueString()
		fmt.Printf("Found metric: name=%s, value=%s\n", metric.ID, val)
		return &metric
	}
	return nil
}

