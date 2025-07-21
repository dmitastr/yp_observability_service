package filestorage

import (
	"time"

	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/dmitastr/yp_observability_service/internal/repository"
)

type FileStorage struct {
	repository.Database
	StoreInterval int
}

func New(db repository.Database, storeInterval int) *FileStorage {
	return &FileStorage{Database: db, StoreInterval: storeInterval}
}


func (fs *FileStorage) SaveData() {
	ticker := time.NewTicker(time.Duration(fs.StoreInterval) * time.Second)
	for range ticker.C {
		if err := fs.Flush(); err != nil {
			logger.GetLogger().Error(err)
		}
	}
}

func (fs *FileStorage) Run() {
	go fs.SaveData()
}