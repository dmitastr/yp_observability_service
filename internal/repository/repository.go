package repository

import models "github.com/dmitastr/yp_observability_service/internal/model"

type Database interface {
	Update(models.Metrics)
	GetAll() []models.Metrics
	Get(key string) *models.Metrics
}