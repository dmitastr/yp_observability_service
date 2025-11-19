package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	grpcapp "github.com/dmitastr/yp_observability_service/internal/app/grpc"
	httpapp "github.com/dmitastr/yp_observability_service/internal/app/http"
	"github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/domain/audit"
	"github.com/dmitastr/yp_observability_service/internal/domain/audit/listener"
	"github.com/dmitastr/yp_observability_service/internal/domain/pinger/postgres_pinger"
	"github.com/dmitastr/yp_observability_service/internal/domain/service"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	dbinterface "github.com/dmitastr/yp_observability_service/internal/repository"
	"github.com/dmitastr/yp_observability_service/internal/repository/filestorage"
	db "github.com/dmitastr/yp_observability_service/internal/repository/memstorage"
	"github.com/dmitastr/yp_observability_service/internal/repository/postgres_storage"
)

type App struct {
	httpServer *http.Server
	gRPCServer *grpcapp.App
	db         dbinterface.Database
}

// NewApp creates a new app, register all handlers and middleware
// and inject necessary dependencies
func NewApp(ctx context.Context, cfg *serverenvconfig.Config) (*App, error) {
	var storage dbinterface.Database
	var err error

	if cfg.DBUrl == nil || *cfg.DBUrl == "" {
		fileStorage := filestorage.New(cfg)
		storage = db.NewStorage(cfg, fileStorage)
	} else {
		storage, err = postgresstorage.NewPG(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("error creating postgres storage: %w", err)
		}
	}

	if err := storage.Init("file://migrations"); err != nil {
		return nil, fmt.Errorf("error initializing postgres storage: %w", err)
	}

	pinger := postgrespinger.New()
	auditor := audit.NewAuditor().
		AddListener(listener.NewListener(listener.FileListenerType, cfg.AuditFile)).
		AddListener(listener.NewListener(listener.URLListenerType, cfg.AuditURL))

	observabilityService := service.NewService(storage, pinger, auditor)

	server, err := httpapp.NewServer(ctx, cfg, observabilityService)
	if err != nil {
		return nil, fmt.Errorf("error creating server: %w", err)
	}

	app := &App{
		httpServer: server,
		db:         storage,
	}

	if cfg.GRPCAddress != nil && *cfg.GRPCAddress != "" {
		app.gRPCServer = grpcapp.NewApp(*cfg.GRPCAddress)
	}

	return app, nil
}

func (app *App) RunHTTPServer() error {
	logger.Infof("Starting app on address: %s\n", app.httpServer.Addr)

	if err := app.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("app error: %w", err)
	}
	return nil
}

func (app *App) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return errors.Join(app.db.Close(), app.httpServer.Shutdown(shutdownCtx))
}

func (app *App) HasgRPCServer() bool {
	return app.gRPCServer != nil
}

func (app *App) RunGRPCServer() error {
	logger.Info("Starting gRPC server")

	app.gRPCServer.MustRun()
	return nil
}

func (app *App) ShutdownGRPCServer() error {
	logger.Info("Stopping gRPC server")
	app.gRPCServer.Stop()
	return nil
}
