package filestorage

import (
	"encoding/json"
	"os"
	"time"

	serverenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	models "github.com/dmitastr/yp_observability_service/internal/model"
	dbinterface "github.com/dmitastr/yp_observability_service/internal/repository/database"
)

type FileStorage struct {
	StoreInterval int
	FileName      string
}

func New(cfg serverenvconfig.Config, db dbinterface.Database) *FileStorage {
	return &FileStorage{StoreInterval: *cfg.StoreInterval, FileName: *cfg.FileStoragePath}
}

func (fs *FileStorage) SaveData(fnc func() []models.Metrics) {
	ticker := time.NewTicker(time.Duration(fs.StoreInterval) * time.Second)
	for range ticker.C {
		metrics := fnc()
		if err := fs.Flush(metrics); err != nil {
			logger.GetLogger().Errorf("Error while saving data to a file: %v", err)
		}
	}
}

func (fs *FileStorage) RunBackup(fnc func() []models.Metrics) {
	go fs.SaveData(fnc)
}

func (fs *FileStorage) createFile() *os.File {
	file, err := os.Create(fs.FileName)
	if err != nil {
		logger.GetLogger().Infof("error while creating file '%s': %v", fs.FileName, err)
	}
	return file
}

func (fs *FileStorage) Flush(metrics []models.Metrics) error {
	file := fs.createFile()
	defer file.Close()
	if err := json.NewEncoder(file).Encode(metrics); err != nil {
		logger.GetLogger().Error(err)
		return err
	}
	return nil
}

func (fs *FileStorage) Load() (metrics []models.Metrics, err error) {
	file, err := os.Open(fs.FileName)
	if err != nil {
		logger.GetLogger().Infof("error while opening file '%s': %s", fs.FileName, err)
		return
	}
	defer file.Close()

	if err = json.NewDecoder(file).Decode(&metrics); err != nil {
		logger.GetLogger().Error(err)
		return
	}
	return
}
