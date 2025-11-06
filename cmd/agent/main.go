package main

import (
	"github.com/dmitastr/yp_observability_service/internal/agent/client"
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

	if err := client.Run(); err != nil {
		logger.Fatal(err)
	}
}
