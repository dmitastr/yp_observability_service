package main

import (
	_ "net/http/pprof"

	"github.com/dmitastr/yp_observability_service/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		panic(err)
	}
}
