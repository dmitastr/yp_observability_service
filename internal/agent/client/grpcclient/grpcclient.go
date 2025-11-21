package grpcclient

import (
	"fmt"
	"os"

	"github.com/dmitastr/yp_observability_service/internal/agent/models"
	"github.com/dmitastr/yp_observability_service/internal/common"
	config "github.com/dmitastr/yp_observability_service/internal/config/env_parser/agent/agent_env_config"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/dmitastr/yp_observability_service/internal/proto/genproto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	client genproto.MetricsClient
	conn   *grpc.ClientConn
}

func NewClient(cfg config.Config) (*Client, error) {
	conn, err := grpc.NewClient(*cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatalf("could not create grpc client: %v", err)
		os.Exit(1)
	}

	c := genproto.NewMetricsClient(conn)
	return &Client{client: c, conn: conn}, nil
}

func (c Client) SendMetric(ctx context.Context, m models.Metric) error {
	return c.SendMetricsBatch(ctx, []models.Metric{m})
}

func (c Client) SendMetricsBatch(ctx context.Context, metrics []models.Metric) error {
	ipValue := ctx.Value(models.RealIP{})

	var ip string
	ip, ok := ipValue.(string)
	if !ok {
		ip = ""
	}

	var metricsProto []*genproto.Metric
	for _, metric := range metrics {
		m := &genproto.Metric_builder{Id: metric.GetID()}

		switch mType := metric.GetMType(); mType {
		case common.COUNTER:
			m.Delta = metric.GetValue().(int64)
			m.Type = genproto.Metric_COUNTER
		case common.GAUGE:
			m.Value = metric.GetValue().(float64)
			m.Type = genproto.Metric_GAUGE
		}
		metricsProto = append(metricsProto, m.Build())
	}

	md := metadata.Pairs("x-real-ip", ip)
	ctx = metadata.NewOutgoingContext(ctx, md)

	in := genproto.UpdateMetricsRequest_builder{Metrics: metricsProto}.Build()

	_, err := c.client.UpdateMetrics(ctx, in)
	if err != nil {
		return fmt.Errorf("could not send metrics: %v", err)
	}
	return nil
}

func (c Client) Close(ctx context.Context) error {
	return c.conn.Close()
}
