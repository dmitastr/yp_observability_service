package client

import (
	"github.com/dmitastr/yp_observability_service/internal/agent/client/grpcclient"
	"github.com/dmitastr/yp_observability_service/internal/agent/client/httpclient"
	model "github.com/dmitastr/yp_observability_service/internal/agent/models"
	config "github.com/dmitastr/yp_observability_service/internal/config/env_parser/agent/agent_env_config"
	"golang.org/x/net/context"
)

type Client interface {
	SendMetric(context.Context, model.Metric) error
	SendMetricsBatch(context.Context, []model.Metric) error
	Close(context.Context) error
}

func NewClient(cfg config.Config) (Client, error) {
	if *cfg.GRPCEnable {
		return grpcclient.NewClient(cfg)
	}
	return httpclient.NewClient(cfg)
}
