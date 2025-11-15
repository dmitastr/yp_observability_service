package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmitastr/yp_observability_service/internal/app"
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

	server, db, err := app.NewApp(ctx)
	if err != nil {
		logger.Fatal(err)
	}

	g, gCtx := errgroup.WithContext(ctx)
	// Server goroutine
	g.Go(func() error {
		logger.Infof("Starting app on address: %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("app error: %w", err)
		}
		logger.Info("Server stopped")
		return nil
	})

	// Server shutdown goroutine
	g.Go(func() error {
		<-gCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	})

	// Database closing goroutine
	g.Go(func() error {
		<-gCtx.Done()
		return db.Close()
	})

	if err := g.Wait(); err != nil {
		logger.Infof("exit reason: %v", err)
	}
}
