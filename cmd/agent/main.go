package main

import (
	"github.com/dmitastr/yp_observability_service/internal/agent/client"
)

func main() {
	agent := client.NewAgent()
	agent.Run(2, 10)
}
