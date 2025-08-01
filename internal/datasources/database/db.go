package database

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	serverenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/errs"
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
	FileName    string
	StreamWrite bool
}

func NewStorage(cfg serverenvconfig.Config) *Storage {
	storage := Storage{FileName: *cfg.FileStoragePath, Metrics: make(map[string]models.Metrics)}
	if *cfg.StoreInterval == 0 {
		storage.StreamWrite = true
	}

	if *cfg.Restore {
		err := storage.Load()
		if err != nil {
			logger.GetLogger().Error(err)
		}
	}

	return &storage
}

func (storage *Storage) CreateFile() *os.File {
	file, err := os.Create(storage.FileName)
	if err != nil {
		logger.GetLogger().Panicf("error while creating file '%s': %s", storage.FileName, err)
	}
	return file
}

func (storage *Storage) OpenFile() *os.File {
	file, err := os.Open(storage.FileName)
	if err != nil {
		logger.GetLogger().Panicf("error while opening file '%s': %s", storage.FileName, err)
	}
	return file
}

func (storage *Storage) Flush() error {
	storage.Lock()
	defer storage.Unlock()
	file := storage.CreateFile()
	if err := json.NewEncoder(file).Encode(storage.toList()); err != nil {
		logger.GetLogger().Error(err)
		return err
	}
	return nil
}

func (storage *Storage) Load() error {
	storage.Lock()
	defer storage.Unlock()
	file, err := os.Open(storage.FileName)
	if err != nil {
		logger.GetLogger().Error("error while opening file '%s': %s", storage.FileName, err)
		return err
	}

	metrics := make([]models.Metrics, 0)
	if err := json.NewDecoder(file).Decode(&metrics); err != nil {
		logger.GetLogger().Fatal(err)
		return err
	}
	storage.Metrics = storage.fromList(metrics)
	return nil
}

func (storage *Storage) fromList(metrics []models.Metrics) map[string]models.Metrics {
	mapping := make(map[string]models.Metrics)
	for _, metric := range metrics {
		mapping[metric.ID] = metric
	}
	return mapping
}

func (storage *Storage) Update(ctx context.Context, newMetric models.Metrics) error {
	logger.GetLogger().Infof("Get new data: %s", newMetric.String())
	storage.Lock()
	defer storage.Unlock()
	if metric, ok := storage.Metrics[newMetric.ID]; ok {
		if metric.Delta != nil {
			newMetric.DeltaSet(metric.Delta)
		}
	}
	storage.Metrics[newMetric.ID] = newMetric
	if storage.StreamWrite {
		if err := storage.Flush(); err != nil {
			logger.GetLogger().Error(err)
		}
	}
	return nil
}

func (storage *Storage) GetAll(ctx context.Context) ([]models.Metrics, error) {
	return storage.toList(), nil
}

func (storage *Storage) Get(ctx context.Context, key string) (*models.Metrics, error) {
	if metric, ok := storage.Metrics[key]; ok {
		logger.GetLogger().Infof("Found metric: %s", metric)
		return &metric, nil
	}
	return nil, errs.ErrorMetricDoesNotExist
}

func (storage *Storage) toList() (lst []models.Metrics) {
	for _, metric := range storage.Metrics {
		lst = append(lst, metric)
	}
	return
}

func (storage *Storage) Close() error {
	return storage.Flush()
}

func (storage *Storage) Ping(ctx context.Context) error {
	return nil
}