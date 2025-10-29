package repository

import (
	"github.com/dmitastr/yp_observability_service/internal/domain/models"
)

type BackupManager interface {
	RunBackup(func() []models.Metrics)
	Load() ([]models.Metrics, error)
	Flush([]models.Metrics) error
}
