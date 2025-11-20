package agent

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/dmitastr/yp_observability_service/internal/agent/client"
	"github.com/dmitastr/yp_observability_service/internal/agent/models"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"golang.org/x/sync/errgroup"

	config "github.com/dmitastr/yp_observability_service/internal/config/env_parser/agent/agent_env_config"
	"github.com/dmitastr/yp_observability_service/internal/errs"
	"github.com/dmitastr/yp_observability_service/internal/logger"
)

const (
	Alloc         = "Alloc"
	BuckHashSys   = "BuckHashSys"
	Frees         = "Frees"
	GCCPUFraction = "GCCPUFraction"
	GCSys         = "GCSys"
	HeapAlloc     = "HeapAlloc"
	HeapIdle      = "HeapIdle"
	HeapInuse     = "HeapInuse"
	HeapObjects   = "HeapObjects"
	HeapReleased  = "HeapReleased"
	HeapSys       = "HeapSys"
	LastGC        = "LastGC"
	Lookups       = "Lookups"
	MCacheInuse   = "MCacheInuse"
	MCacheSys     = "MCacheSys"
	MSpanInuse    = "MSpanInuse"
	MSpanSys      = "MSpanSys"
	Mallocs       = "Mallocs"
	NextGC        = "NextGC"
	NumForcedGC   = "NumForcedGC"
	NumGC         = "NumGC"
	OtherSys      = "OtherSys"
	PauseTotalNs  = "PauseTotalNs"
	StackInuse    = "StackInuse"
	StackSys      = "StackSys"
	Sys           = "Sys"
	TotalAlloc    = "TotalAlloc"

	RandomValue = "RandomValue"
	PollCount   = "PollCount"

	TotalMemory     = "TotalMemory"
	FreeMemory      = "FreeMemory"
	CPUutilization1 = "CPUUtilization1"
)

type Result struct {
	err error
}

type Agent struct {
	sync.Mutex
	Metrics   map[string]models.Metric
	Client    client.Client
	address   string
	RateLimit int
	realAddr  net.Addr
}

func NewAgent(cfg config.Config, client client.Client) (*Agent, error) {
	a := Agent{
		Metrics: make(map[string]models.Metric),
		Client:  client,
	}

	if cfg.RateLimit != nil {
		a.RateLimit = *cfg.RateLimit
	}

	if err := a.setRealIP(); err != nil {
		return nil, fmt.Errorf("error setting realIP: %w", err)
	}

	return &a, nil
}

func (agent *Agent) UpdateMetricValueCounter(key string, value int64) {
	if _, ok := agent.Metrics[key]; !ok {
		pc := models.NewCounterMetric(key, 0)
		agent.Metrics[key] = pc
	}
	pc := agent.Metrics[key]
	pc.UpdateValue(value)
}

func (agent *Agent) UpdateMetricValueGauge(key string, value float64) {
	if _, ok := agent.Metrics[key]; !ok {
		pc := models.NewGaugeMetric(key, 0)
		agent.Metrics[key] = pc
	}
	pc := agent.Metrics[key]
	pc.UpdateValue(value)
}

func (agent *Agent) UpdateMetrics() error {
	agent.Mutex.Lock()
	defer agent.Mutex.Unlock()

	// Update runtime metrics
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	agent.UpdateMetricValueGauge(Alloc, float64(stats.Alloc))
	agent.UpdateMetricValueGauge(BuckHashSys, float64(stats.BuckHashSys))
	agent.UpdateMetricValueGauge(Frees, float64(stats.Frees))
	agent.UpdateMetricValueGauge(GCCPUFraction, stats.GCCPUFraction)
	agent.UpdateMetricValueGauge(GCSys, float64(stats.GCSys))
	agent.UpdateMetricValueGauge(HeapAlloc, float64(stats.HeapAlloc))
	agent.UpdateMetricValueGauge(HeapIdle, float64(stats.HeapIdle))
	agent.UpdateMetricValueGauge(HeapInuse, float64(stats.HeapInuse))
	agent.UpdateMetricValueGauge(HeapObjects, float64(stats.HeapObjects))
	agent.UpdateMetricValueGauge(HeapReleased, float64(stats.HeapReleased))
	agent.UpdateMetricValueGauge(HeapSys, float64(stats.HeapSys))
	agent.UpdateMetricValueGauge(LastGC, float64(stats.LastGC))
	agent.UpdateMetricValueGauge(Lookups, float64(stats.Lookups))
	agent.UpdateMetricValueGauge(MCacheInuse, float64(stats.MCacheInuse))
	agent.UpdateMetricValueGauge(MCacheSys, float64(stats.MCacheSys))
	agent.UpdateMetricValueGauge(MSpanInuse, float64(stats.MSpanInuse))
	agent.UpdateMetricValueGauge(MSpanSys, float64(stats.MSpanSys))
	agent.UpdateMetricValueGauge(Mallocs, float64(stats.Mallocs))
	agent.UpdateMetricValueGauge(NextGC, float64(stats.NextGC))
	agent.UpdateMetricValueGauge(NumForcedGC, float64(stats.NumForcedGC))
	agent.UpdateMetricValueGauge(NumGC, float64(stats.NumGC))
	agent.UpdateMetricValueGauge(OtherSys, float64(stats.OtherSys))
	agent.UpdateMetricValueGauge(PauseTotalNs, float64(stats.PauseTotalNs))
	agent.UpdateMetricValueGauge(StackInuse, float64(stats.StackInuse))
	agent.UpdateMetricValueGauge(StackSys, float64(stats.StackSys))
	agent.UpdateMetricValueGauge(Sys, float64(stats.Sys))
	agent.UpdateMetricValueGauge(TotalAlloc, float64(stats.TotalAlloc))
	agent.UpdateMetricValueGauge(RandomValue, 100*rand.Float64())

	agent.UpdateMetricValueCounter(PollCount, 1)

	// Update sys utils metrics
	v, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("error getting virtual memory info: %w", err)
	}

	agent.UpdateMetricValueGauge(TotalMemory, float64(v.Total))
	agent.UpdateMetricValueGauge(FreeMemory, float64(v.Free))
	cpuStats, _ := cpu.Info()
	if len(cpuStats) > 0 {
		agent.UpdateMetricValueGauge(CPUutilization1, float64(cpuStats[0].CPU))
	}
	return nil
}

func (agent *Agent) addIP(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, models.RealIP{}, agent.realAddr)
	return ctx
}

func (agent *Agent) SendMetric(ctx context.Context, key string) error {
	ctx = agent.addIP(ctx)
	metric, ok := agent.Metrics[key]
	if !ok {
		return errs.ErrorMetricDoesNotExist
	}
	if err := agent.Client.SendMetric(ctx, metric); err != nil {
		return fmt.Errorf("error sending metric: %w", err)
	}

	return nil
}

func (agent *Agent) SendMetricsBatch(ctx context.Context, metrics []models.Metric) error {
	ctx = agent.addIP(ctx)
	if err := agent.Client.SendMetricsBatch(ctx, metrics); err != nil {
		return fmt.Errorf("error sending metrics batch: %w", err)
	}

	return nil
}

func (agent *Agent) toList() (metrics []models.Metric) {
	for _, metric := range agent.Metrics {
		metrics = append(metrics, metric)
	}
	return
}

func (agent *Agent) calculateChunkSize() int {
	return int(math.Ceil(float64(len(agent.Metrics)) / float64(agent.RateLimit)))
}

func (agent *Agent) FeedWorkers(inCh chan []models.Metric) {
	metrics := agent.toList()
	chunkSize := agent.calculateChunkSize()

	for i := 0; i < len(metrics); i += chunkSize {
		end := i + chunkSize
		if end > len(metrics) {
			end = len(metrics)
		}
		inCh <- metrics[i:end]
	}
}

func (agent *Agent) startWorkers(ctx context.Context, inCh chan []models.Metric) error {
	g, gCtx := errgroup.WithContext(ctx)
	for w := range agent.RateLimit {
		logger.Infof("Worker %d starting", w)
		g.Go(func() error {
			for {
				select {
				case <-gCtx.Done():
					logger.Infof("Worker %d shutting down", w)
					return nil

				case batch := <-inCh:
					if err := agent.SendMetricsBatch(ctx, batch); err != nil {
						return fmt.Errorf("error sending metrics batch: %w", err)
					}
				}
			}
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error in workers goroutine: %w", err)
	}
	return nil
}

func (agent *Agent) Run(ctx context.Context, pollInterval int, reportInterval int) error {
	batchCh := make(chan []models.Metric)

	g, gCtx := errgroup.WithContext(ctx)

	// Start workers to read batches of data
	g.Go(func() error {
		return agent.startWorkers(ctx, batchCh)
	})

	// Create batches of data
	g.Go(func() error {
		ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-gCtx.Done():
				logger.Info("Shutting down report data goroutine")
				return nil
			case <-ticker.C:
				agent.FeedWorkers(batchCh)
			}
		}
	})

	// Collect stats
	g.Go(func() error {
		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-gCtx.Done():
				logger.Info("Shutting down report data goroutine")
				return nil
			case <-ticker.C:
				if err := agent.UpdateMetrics(); err != nil {
					return fmt.Errorf("failed to update metrics: %w", err)
				}
			}
		}
	})

	return g.Wait()
}

func (agent *Agent) setRealIP() error {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	agent.realAddr = localAddr

	return nil
}

func (agent *Agent) Stop(ctx context.Context) error {
	return agent.Client.Close(ctx)
}
