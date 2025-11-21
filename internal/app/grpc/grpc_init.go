package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/dmitastr/yp_observability_service/internal/domain/models"
	"github.com/dmitastr/yp_observability_service/internal/domain/service"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/dmitastr/yp_observability_service/internal/presentation/middleware/ipchecker"
	"github.com/dmitastr/yp_observability_service/internal/proto/genproto"
	"google.golang.org/grpc"
)

type App struct {
	address     string
	gRPCServer  *grpc.Server
	ipValidator *ipchecker.IPValidator
}

func NewApp(address string, observerService service.IService, ipValidator *ipchecker.IPValidator) *App {
	server := NewServer(observerService, ipValidator)
	return &App{address: address, gRPCServer: server, ipValidator: ipValidator}
}

// MustRun runs gRPC server and panics if any error occurs
func (app *App) MustRun() {
	if err := app.run(); err != nil {
		panic(err)
	}
}

// Stop stops gRPC server
func (app *App) Stop() {
	logger.Info("stopping gRPC server")
	app.gRPCServer.GracefulStop()
}

// run starts listening on address
func (app *App) run() error {
	lis, err := net.Listen("tcp", app.address)
	if err != nil {
		return fmt.Errorf("cannot listen: %w", err)
	}

	if err := app.gRPCServer.Serve(lis); err != nil {
		return fmt.Errorf("cannot serve: %w", err)
	}
	return nil
}

// MetricsServer handles gRPC requests
type MetricsServer struct {
	genproto.UnimplementedMetricsServer
	service service.IService
}

func (m MetricsServer) UpdateMetrics(ctx context.Context, in *genproto.UpdateMetricsRequest) (*genproto.UpdateMetricsResponse, error) {
	var response genproto.UpdateMetricsResponse

	var metrics []models.Metrics
	for _, mProto := range in.GetMetrics() {
		metric := models.FromProto(mProto)
		metrics = append(metrics, metric)
	}

	if err := m.service.BatchUpdate(ctx, metrics); err != nil {
		return nil, fmt.Errorf("error updating metrics: %w", err)
	}

	return &response, nil
}

func NewServer(observerService service.IService, ipValidator *ipchecker.IPValidator) *grpc.Server {
	s := grpc.NewServer(
		grpc.UnaryInterceptor(ipValidator.IPValidatorInterceptor),
	)
	genproto.RegisterMetricsServer(s, MetricsServer{service: observerService})
	return s
}
