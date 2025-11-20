package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	agent "github.com/dmitastr/yp_observability_service/internal/agent/init"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"golang.org/x/sync/errgroup"
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

	g, gCtx := errgroup.WithContext(ctx)
	// Agent start goroutine
	g.Go(func() error {
		if err := agent.Run(ctx); err != nil {
			return fmt.Errorf("agent error: %w", err)
		}
		logger.Info("Agent stopped")
		return nil
	})

	// Agent shutdown goroutine
	g.Go(func() error {
		<-gCtx.Done()
		return agent.Stop(gCtx)
	})

	if err := g.Wait(); err != nil {
		logger.Infof("exit reason: %v", err)
	}

}
