package logger

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
)

var sLog *ZapSugarLogger

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
	sync.OnceFunc(func() {
		cfg := zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

		logger, err := cfg.Build()
		if err != nil {
			panic(err)
		}

		sLog = &ZapSugarLogger{SugaredLogger: logger.Sugar()}
	})()
}

func GetLogger() *ZapSugarLogger {
	if sLog == nil {
		Initialize()
		return sLog
	}
	return sLog
}

func getLogger() *ZapSugarLogger {
	if sLog == nil {
		Initialize()
		return sLog
	}
	return sLog
}

func Info(msg string) {
	getLogger().Info(msg)
}
func Infof(msg string, args ...any) {
	getLogger().Infof(msg, args...)
}
func Warn(msg string) {
	getLogger().Warn(msg)
}
func Warnf(msg string, args ...any) {
	getLogger().Warnf(msg, args...)
}
func Error(err error) {
	getLogger().Error(err)
}
func Errorf(msg string, args ...any) {
	getLogger().Errorf(msg, args...)
}
func Fatal(args ...any) {
	getLogger().Fatal(args...)
}
func Fatalf(msg string, args ...any) {
	getLogger().Fatalf(msg, args...)
}
func Panic(msg string) {
	getLogger().Panic(msg)
}
func Panicf(msg string, args ...any) {
	getLogger().Panicf(msg, args...)
}
