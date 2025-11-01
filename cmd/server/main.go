package main

import (
	_ "net/http/pprof"

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

	if err := server.Run(); err != nil {
		panic(err)
	}
}
