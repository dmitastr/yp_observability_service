package logger

import (
	"go.uber.org/zap"
)

var SLog *zap.SugaredLogger

func Initialize() {
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	SLog = logger.Sugar()
}

func GetLogger() *zap.SugaredLogger {
	if SLog == nil {
		Initialize()
	}
	return SLog
}