package backupmanager

import models "github.com/dmitastr/yp_observability_service/internal/model"

type BackupManager interface {
	RunBackup(func() []models.Metrics)
	Load() ([]models.Metrics, error)
	Flush([]models.Metrics) error
}
