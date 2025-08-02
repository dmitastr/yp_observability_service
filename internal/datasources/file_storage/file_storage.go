package filestorage

import (
	"encoding/json"
	"os"

	serverenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	models "github.com/dmitastr/yp_observability_service/internal/model"
)

type FileStorage struct {
	StoreInterval int
	FileName    string
}

func New(cfg serverenvconfig.Config) *FileStorage {
	return &FileStorage{StoreInterval: *cfg.StoreInterval, FileName: *cfg.FileStoragePath}
}



func (fs *FileStorage) CreateFile() *os.File {
	file, err := os.Create(fs.FileName)
	if err != nil {
		logger.GetLogger().Infof("error while creating file '%s': %v", fs.FileName, err)
	}
	return file
}

func (fs *FileStorage) OpenFile() *os.File {
	file, err := os.Open(fs.FileName)
	if err != nil {
		logger.GetLogger().Infof("error while opening file '%s': %v", fs.FileName, err)
	}
	return file
}

func (fs *FileStorage) Flush(metrics []models.Metrics) error {
	file := fs.CreateFile()
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

	if err = json.NewDecoder(file).Decode(&metrics); err != nil {
		logger.GetLogger().Fatal(err)
		return 
	}
	return 
}