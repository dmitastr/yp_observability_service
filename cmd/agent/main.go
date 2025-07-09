package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/dmitastr/yp_observability_service/internal/agent/client"
)


var serverAddress string
var reportInterval int
var pollInterval int

func init() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "set server host and port")
	flag.IntVar(&reportInterval, "r", 10, "frequency of data sending to server in seconds")
	flag.IntVar(&pollInterval, "p", 2, "frequency of metric polling from source in seconds")
}

func configParse() {
	if val, ok := os.LookupEnv("ADDRESS"); ok {
		serverAddress = val
	}
	if val, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		parsedVal, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}
		pollInterval = parsedVal
	}
	if val, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		parsedVal, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}
		reportInterval = parsedVal
	}
}

func main() {
	flag.Parse()
	configParse()

	fmt.Printf("Starting client for server=%s, poll interval=%d, report interval=%d\n", serverAddress, pollInterval, reportInterval)

	agent := client.NewAgent(serverAddress)
	agent.Run(pollInterval, reportInterval)
}
