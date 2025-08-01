package main

import (
	"flag"
	"net/http"

	"github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/dmitastr/yp_observability_service/internal/server"
)

var ServerAddress string
var StoreInterval int
var Restore bool
var FileStoragePath string
var DBUrl string

func init() {
	flag.StringVar(&ServerAddress, "a", "localhost:8080", "set server host and port")
	flag.IntVar(&StoreInterval, "i", 300, "interval for storing data to the file in seconds, 0=stream writing")
	flag.BoolVar(&Restore, "r", false, "restore data from file")
	flag.StringVar(&FileStoragePath, "f", "data/data.json", "path for writing data")
	flag.StringVar(&DBUrl, "d", "", "database connection url")
}

func main() {
	flag.Parse()
	logger.Initialize()
	cfg := serverenvconfig.New(ServerAddress, StoreInterval, FileStoragePath, Restore, DBUrl)

	router, db := server.NewServer(cfg)
	defer db.Close()

	logger.GetLogger().Infof("Starting server=%s\n", *cfg.Address)
	if err := http.ListenAndServe(*cfg.Address, router); err != nil {
		panic(err)
	}
}
