package main

import (
	"context"
	_ "net/http/pprof"
	"os"
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

	ctx, cancel := context.WithCancel(context.Background())

	cancelCh := make(chan os.Signal, 1)
	go func() {
		signal.Notify(cancelCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
		<-cancelCh
		logger.Info("Received an interrupt, shutting down...")
		cancel()
	}()

	if err := server.Run(ctx); err != nil {
		panic(err)
	}
}
