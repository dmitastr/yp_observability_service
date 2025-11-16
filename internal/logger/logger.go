package logger

import (
	"context"
	"fmt"
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

func Initialize() error {
	var err error
	sync.OnceFunc(func() {
		cfg := zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

		logger, errLogger := cfg.Build(zap.AddCallerSkip(1))
		if errLogger != nil {
			err = fmt.Errorf("could not initialize zap logger: %w", errLogger)
		}

		sLog = &ZapSugarLogger{SugaredLogger: logger.Sugar()}
	})()
	return err
}

func GetLogger() *ZapSugarLogger {
	if sLog == nil {
		_ = Initialize()
		return sLog
	}
	return sLog
}

func Info(msg string) {
	GetLogger().Info(msg)
}
func Infof(msg string, args ...any) {
	GetLogger().Infof(msg, args...)
}
func Warn(msg string) {
	GetLogger().Warn(msg)
}
func Warnf(msg string, args ...any) {
	GetLogger().Warnf(msg, args...)
}
func Error(err error) {
	GetLogger().Error(err)
}
func Errorf(msg string, args ...any) {
	GetLogger().Errorf(msg, args...)
}
func Fatal(args ...any) {
	GetLogger().Fatal(args...)
}
func Fatalf(msg string, args ...any) {
	GetLogger().Fatalf(msg, args...)
}
func Panic(msg string) {
	GetLogger().Panic(msg)
}
func Panicf(msg string, args ...any) {
	GetLogger().Panicf(msg, args...)
}
