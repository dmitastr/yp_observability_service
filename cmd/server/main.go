package main

import (
	_ "net/http/pprof"

	"github.com/dmitastr/yp_observability_service/internal/server"
)

func main() {
	if err := server.Execute(); err != nil {
		panic(err)
	}
}
