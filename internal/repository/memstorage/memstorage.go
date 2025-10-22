package memstorage

import (
	"context"
	"fmt"
	"sync"

	serverenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/domain/models"
	"github.com/dmitastr/yp_observability_service/internal/errs"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	backupmanager "github.com/dmitastr/yp_observability_service/internal/repository"
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
	Metrics       map[string]models.Metrics
	BackupManager backupmanager.BackupManager
	StreamWrite   bool
	Restore       bool
}

func NewStorage(cfg serverenvconfig.Config, bm backupmanager.BackupManager) *Storage {
	storage := Storage{Metrics: make(map[string]models.Metrics)}
	if *cfg.StoreInterval == 0 {
		storage.StreamWrite = true
	}
	storage.Restore = *cfg.Restore
	storage.BackupManager = bm

	return &storage
}

func (storage *Storage) Init() error {
	if !storage.StreamWrite {
		storage.BackupManager.RunBackup(storage.toList)
	}

	if storage.Restore {
		metrics, err := storage.BackupManager.Load()
		if err != nil {
			return fmt.Errorf("error loading metrics from file backup: %w", err)
		}
		storage.Metrics = storage.fromList(metrics)
	}

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
	logger.Infof("Get new data: %s", newMetric.String())
	storage.Lock()
	defer storage.Unlock()
	if metric, ok := storage.Metrics[newMetric.ID]; ok {
		if metric.Delta != nil {
			newMetric.UpdateDelta(metric.Delta)
		}
	}
	storage.Metrics[newMetric.ID] = newMetric
	if storage.StreamWrite {
		metrics := storage.toList()
		if err := storage.BackupManager.Flush(metrics); err != nil {
			logger.Error(err)
		}
	}
	return nil
}

func (storage *Storage) BulkUpdate(ctx context.Context, metrics []models.Metrics) error {
	logger.Infof("Get %d new metrics", len(metrics))
	for _, metric := range metrics {
		err := storage.Update(ctx, metric)
		if err != nil {
			return err
		}
	}

	if storage.StreamWrite {
		metrics := storage.toList()
		if err := storage.BackupManager.Flush(metrics); err != nil {
			return err
		}

	}
	return nil
}

func (storage *Storage) GetAll(ctx context.Context) ([]models.Metrics, error) {
	return storage.toList(), nil
}

func (storage *Storage) Get(ctx context.Context, key string) (*models.Metrics, error) {
	if metric, ok := storage.Metrics[key]; ok {
		logger.Infof("Found metric: %s", metric)
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
	if err := storage.BackupManager.Flush(metrics); err != nil {
		return fmt.Errorf("error saving metrics to file backup: %w", err)
	}
	return nil
}

func (storage *Storage) Ping(ctx context.Context) error {
	return nil
}
