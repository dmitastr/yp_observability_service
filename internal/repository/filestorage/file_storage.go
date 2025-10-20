package filestorage

import (
	"encoding/json"
	"os"
	"time"

	serverenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/domain/model"
	"github.com/dmitastr/yp_observability_service/internal/logger"
)

type FileStorage struct {
	StoreInterval int
	FileName      string
}

func New(cfg serverenvconfig.Config) *FileStorage {
	return &FileStorage{StoreInterval: *cfg.StoreInterval, FileName: *cfg.FileStoragePath}
}

func (fs *FileStorage) SaveData(fnc func() []models.Metrics) {
	ticker := time.NewTicker(time.Duration(fs.StoreInterval) * time.Second)
	for range ticker.C {
		metrics := fnc()
		if err := fs.Flush(metrics); err != nil {
			logger.Errorf("Error while saving data to a file: %v", err)
		}
	}
}

func (fs *FileStorage) RunBackup(fnc func() []models.Metrics) {
	go fs.SaveData(fnc)
}

func (fs *FileStorage) createFile() *os.File {
	file, err := os.Create(fs.FileName)
	if err != nil {
		logger.Errorf("error while creating file '%s': %v", fs.FileName, err)
	}
	return file
}

func (fs *FileStorage) Flush(metrics []models.Metrics) error {
	file := fs.createFile()
	defer file.Close()
	if err := json.NewEncoder(file).Encode(metrics); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func (fs *FileStorage) Load() (metrics []models.Metrics, err error) {
	file, err := os.Open(fs.FileName)
	if err != nil {
		logger.Infof("error while opening file '%s': %s", fs.FileName, err)
		return metrics, err
	}
	defer file.Close()

	if err = json.NewDecoder(file).Decode(&metrics); err != nil {
		logger.Error(err)
		return
	}
	return
}
