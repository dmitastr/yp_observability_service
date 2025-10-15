package logger

import (
	"context"

	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
)

var SLog *ZapSugarLogger

type ZapSugarLogger struct {
	*zap.SugaredLogger
}

func (zl *ZapSugarLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	if err, ok := data["err"].(error); ok && err != nil {
		data["err"] = err.Error()
	}

	switch level {
	case tracelog.LogLevelDebug:
		zl.Debugw(msg, "pgx_data", data)
	case tracelog.LogLevelInfo:
		zl.Infow(msg, "pgx_data", data)
	case tracelog.LogLevelWarn:
		zl.Warnw(msg, "pgx_data", data)
	case tracelog.LogLevelError:
		zl.Errorw(msg, "pgx_data", data)
	default:
		zl.Infow(msg, "pgx_data", data)
	}
}

func Initialize() {
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	SLog = &ZapSugarLogger{SugaredLogger: logger.Sugar()}
}

func GetLogger() *ZapSugarLogger {
	if SLog == nil {
		Initialize()
		return SLog
	}
	return SLog
}
