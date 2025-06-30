package client

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/dmitastr/yp_observability_service/internal/agent/metric"
)

// Alloc
// BuckHashSys
// Frees
// GCCPUFraction
// GCSys
// HeapAlloc
// HeapIdle
// HeapInuse
// HeapObjects
// HeapReleased
// HeapSys
// LastGC
// Lookups
// MCacheInuse
// MCacheSys
// MSpanInuse
// MSpanSys
// Mallocs
// NextGC
// NumForcedGC
// NumGC
// OtherSys
// PauseTotalNs
// StackInuse
// StackSys
// Sys
// TotalAlloc


type Agent struct {
	Metrics map[string]metric.Metric
	Client  http.Client
	address string
}

func NewAgent(address string) *Agent {
	agent := Agent{
		Metrics: make(map[string]metric.Metric), 
		Client: http.Client{},
		address: address,
	}
	return &agent
}

func (agent *Agent) UpdateMetricValueCounter(key string, value int64) {
	if _, ok := agent.Metrics[key]; !ok {
		pc := metric.NewCounterMetric(key, 0)
		agent.Metrics[key] = pc
	}
	pc := agent.Metrics[key]
	pc.UpdateValue(value)
}

func (agent *Agent) UpdateMetricValueGauge(key string, value float64) {
	if _, ok := agent.Metrics[key]; !ok {
		pc := metric.NewGaugeMetric(key, 0)
		agent.Metrics[key] = pc
	}
	pc := agent.Metrics[key]
	pc.UpdateValue(value)
}

func (agent *Agent) Update(pollInterval int) {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)

		agent.UpdateMetricValueGauge("Alloc", float64(stats.Alloc))
		agent.UpdateMetricValueGauge("BuckHashSys", float64(stats.BuckHashSys))
		agent.UpdateMetricValueGauge("Frees", float64(stats.Frees))
		agent.UpdateMetricValueGauge("GCCPUFraction", float64(stats.GCCPUFraction))
		agent.UpdateMetricValueGauge("GCSys", float64(stats.GCSys))
		agent.UpdateMetricValueGauge("HeapAlloc", float64(stats.HeapAlloc))
		agent.UpdateMetricValueGauge("HeapIdle", float64(stats.HeapIdle))
		agent.UpdateMetricValueGauge("HeapInuse", float64(stats.HeapInuse))
		agent.UpdateMetricValueGauge("HeapObjects", float64(stats.HeapObjects))
		agent.UpdateMetricValueGauge("HeapReleased", float64(stats.HeapReleased))
		agent.UpdateMetricValueGauge("HeapSys", float64(stats.HeapSys))
		agent.UpdateMetricValueGauge("LastGC", float64(stats.LastGC))
		agent.UpdateMetricValueGauge("Lookups", float64(stats.Lookups))
		agent.UpdateMetricValueGauge("MCacheInuse", float64(stats.MCacheInuse))
		agent.UpdateMetricValueGauge("MCacheSys", float64(stats.MCacheSys))
		agent.UpdateMetricValueGauge("MSpanInuse", float64(stats.MSpanInuse))
		agent.UpdateMetricValueGauge("MSpanSys", float64(stats.MSpanSys))
		agent.UpdateMetricValueGauge("Mallocs", float64(stats.Mallocs))
		agent.UpdateMetricValueGauge("NextGC", float64(stats.NextGC))
		agent.UpdateMetricValueGauge("NumForcedGC", float64(stats.NumForcedGC))
		agent.UpdateMetricValueGauge("NumGC", float64(stats.NumGC))
		agent.UpdateMetricValueGauge("OtherSys", float64(stats.OtherSys))
		agent.UpdateMetricValueGauge("PauseTotalNs", float64(stats.PauseTotalNs))
		agent.UpdateMetricValueGauge("StackInuse", float64(stats.StackInuse))
		agent.UpdateMetricValueGauge("StackSys", float64(stats.StackSys))
		agent.UpdateMetricValueGauge("Sys", float64(stats.Sys))
		agent.UpdateMetricValueGauge("TotalAlloc", float64(stats.TotalAlloc))
		agent.UpdateMetricValueGauge("RandomValue", rand.Float64())

		agent.UpdateMetricValueCounter("PollCount", 1)

	}
}

func (agent *Agent) SendMetric(key string) error {
	metric, ok := agent.Metrics[key]
	if !ok {
		return errors.New("metric was not found")
	}
	pathParams := metric.ToString()

	args := make([]any, len(pathParams)+1)
	args[0] = agent.address
	for i := range len(pathParams) {
		args[i+1] = pathParams[i]
	}
	url := fmt.Sprintf(`%s/update/%s/%s/%s`, args...)

	if _, err := agent.Client.Post(url, "text/plain", nil); err != nil {
		return err
	}
	return nil
}

func (agent *Agent) SendData(reportInterval int) {
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		for ID := range agent.Metrics {
			if err := agent.SendMetric(ID); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (agent Agent) Run(pollInterval int, reportInterval int) {
	go agent.Update(pollInterval)
	agent.SendData(reportInterval)
}
