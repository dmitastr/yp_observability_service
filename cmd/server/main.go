package main

import (
	"context"
	_ "net/http/pprof"
	"os/signal"
	"syscall"

	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/dmitastr/yp_observability_service/internal/server"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	logger.Infof("Build version: %s\n", buildVersion)
	logger.Infof("Build data: %s\n", buildDate)
	logger.Infof("Build commit: %s\n", buildCommit)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer func() {
		logger.Info("Received an interrupt, shutting down...")
		stop()
	}()

	if err := server.Run(ctx); err != nil {
		panic(err)
	}
}
