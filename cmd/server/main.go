package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"

	"github.com/dmitastr/yp_observability_service/internal/app"
	serverenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
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

	cfg, err := serverenvconfig.New()
	if err != nil {
		logger.Fatal(err)
	}

	observerApp, err := app.NewApp(ctx, cfg)
	if err != nil {
		logger.Fatal(err)
	}

	g, gCtx := errgroup.WithContext(ctx)

	// Server goroutine
	g.Go(func() error {
		if err := observerApp.RunHTTPServer(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("observerApp error: %w", err)
		}
		logger.Info("Server stopped")
		return nil
	})

	// Server shutdown goroutine
	g.Go(func() error {
		<-gCtx.Done()
		return observerApp.Shutdown(ctx)
	})

	// Starting separate gRPC server
	if observerApp.HasgRPCServer() {
		g.Go(func() error {
			return observerApp.RunGRPCServer()
		})

		g.Go(func() error {
			<-gCtx.Done()
			return observerApp.ShutdownGRPCServer()
		})
	}

	if err := g.Wait(); err != nil {
		logger.Infof("exit reason: %v", err)
	}
}
