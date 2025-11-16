package main

import (
	"context"
	"os/signal"
	"syscall"

	client "github.com/dmitastr/yp_observability_service/internal/agent/init"
	"github.com/dmitastr/yp_observability_service/internal/logger"
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

	if err := client.Run(ctx); err != nil {
		logger.Fatal(err)
	}
}
