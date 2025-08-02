package database

import (
	"context"
	"sync"
	"time"

	serverenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	filestorage "github.com/dmitastr/yp_observability_service/internal/datasources/file_storage"
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
	FileStorage *filestorage.FileStorage
	StreamWrite bool
}

func NewStorage(cfg serverenvconfig.Config) *Storage {
	fileStorage := filestorage.New(cfg)
	storage := Storage{FileStorage: fileStorage, Metrics: make(map[string]models.Metrics)}
	if *cfg.StoreInterval == 0 {
		storage.StreamWrite = true
	}

	if *cfg.Restore {
		storage.Load()
	}

	return &storage
}


func (storage *Storage) Load() error {
	storage.Lock()
	defer storage.Unlock()
	
	metrics, err := storage.FileStorage.Load()
	if err != nil {
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
		metrics := storage.toList()
		if err := storage.FileStorage.Flush(metrics); err != nil {
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
	metrics := storage.toList()
	return storage.FileStorage.Flush(metrics)
}

func (storage *Storage) Ping(ctx context.Context) error {
	return nil
}


func (storage *Storage) SaveData() {
	ticker := time.NewTicker(time.Duration(storage.FileStorage.StoreInterval) * time.Second)
	for range ticker.C {
		metrics := storage.toList()
		if err := storage.FileStorage.Flush(metrics); err != nil {
			logger.GetLogger().Error(err)
		}
	}
}

func (storage *Storage) RunBackup() {
	go storage.SaveData()
}