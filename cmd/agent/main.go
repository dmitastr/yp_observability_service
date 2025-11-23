package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/dmitastr/yp_observability_service/internal/agent/agent"
	config "github.com/dmitastr/yp_observability_service/internal/config/env_parser/agent/agent_env_config"
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

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal(err)
	}

	metricsAgent, err := agent.NewAgent(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	g, gCtx := errgroup.WithContext(ctx)
	// Agent start goroutine
	g.Go(func() error {
		if err := metricsAgent.Run(ctx, *cfg.PollInterval, *cfg.ReportInterval); err != nil {
			return fmt.Errorf("agent error: %w", err)
		}
		logger.Info("Agent stopped")
		return nil
	})

	// Agent shutdown goroutine
	g.Go(func() error {
		<-gCtx.Done()
		return metricsAgent.Stop(gCtx)
	})

	if err := g.Wait(); err != nil {
		logger.Infof("exit reason: %v", err)
	}

}
