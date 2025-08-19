package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	model "github.com/dmitastr/yp_observability_service/internal/agent/metric"
	"github.com/dmitastr/yp_observability_service/internal/common"
	agentenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/agent/agent_env_config"
	"github.com/dmitastr/yp_observability_service/internal/errs"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/dmitastr/yp_observability_service/internal/signature"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	Alloc         string = "Alloc"
	BuckHashSys   string = "BuckHashSys"
	Frees         string = "Frees"
	GCCPUFraction string = "GCCPUFraction"
	GCSys         string = "GCSys"
	HeapAlloc     string = "HeapAlloc"
	HeapIdle      string = "HeapIdle"
	HeapInuse     string = "HeapInuse"
	HeapObjects   string = "HeapObjects"
	HeapReleased  string = "HeapReleased"
	HeapSys       string = "HeapSys"
	LastGC        string = "LastGC"
	Lookups       string = "Lookups"
	MCacheInuse   string = "MCacheInuse"
	MCacheSys     string = "MCacheSys"
	MSpanInuse    string = "MSpanInuse"
	MSpanSys      string = "MSpanSys"
	Mallocs       string = "Mallocs"
	NextGC        string = "NextGC"
	NumForcedGC   string = "NumForcedGC"
	NumGC         string = "NumGC"
	OtherSys      string = "OtherSys"
	PauseTotalNs  string = "PauseTotalNs"
	StackInuse    string = "StackInuse"
	StackSys      string = "StackSys"
	Sys           string = "Sys"
	TotalAlloc    string = "TotalAlloc"

	RandomValue string = "RandomValue"
	PollCount   string = "PollCount"
)

type Result struct {
	err error
}

type Agent struct {
	Metrics    map[string]model.Metric
	Client     *retryablehttp.Client
	address    string
	HashSigner *signature.HashSigner
	RateLimit  int
}

func NewAgent(cfg agentenvconfig.Config) *Agent {
	client := retryablehttp.NewClient()
	client.HTTPClient.Timeout = time.Millisecond * 300
	client.RetryMax = 3
	client.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		return time.Second * time.Duration(2*attemptNum+1)
	}

	address := *cfg.Address
	if !strings.Contains(address, "http") {
		address = "http://" + address
	}

	agent := Agent{
		Metrics:    make(map[string]model.Metric),
		Client:     client,
		address:    address,
		HashSigner: signature.NewHashSigner(cfg.Key),
		RateLimit:  *cfg.RateLimit,
	}
	return &agent
}

func (agent *Agent) UpdateMetricValueCounter(key string, value int64) {
	if _, ok := agent.Metrics[key]; !ok {
		pc := model.NewCounterMetric(key, 0)
		agent.Metrics[key] = pc
	}
	pc := agent.Metrics[key]
	pc.UpdateValue(value)
}

func (agent *Agent) UpdateMetricValueGauge(key string, value float64) {
	if _, ok := agent.Metrics[key]; !ok {
		pc := model.NewGaugeMetric(key, 0)
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

		agent.UpdateMetricValueGauge(Alloc, float64(stats.Alloc))
		agent.UpdateMetricValueGauge(BuckHashSys, float64(stats.BuckHashSys))
		agent.UpdateMetricValueGauge(Frees, float64(stats.Frees))
		agent.UpdateMetricValueGauge(GCCPUFraction, float64(stats.GCCPUFraction))
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

	}
}

func (agent *Agent) Post(url string, data []byte, compressed bool) (resp *http.Response, err error) {
	var postData bytes.Buffer
	var compression string

	if compressed {
		gw := gzip.NewWriter(&postData)
		if _, err := gw.Write(data); err != nil {
			return nil, err
		}
		if err := gw.Close(); err != nil {
			logger.GetLogger().Errorf("failed to close gzip writer: %v", err)
		}
		compression = "gzip"
	} else {
		if _, err := postData.Write(data); err != nil {
			logger.GetLogger().Errorf("failed to write uncompressed: %v", err)
			return nil, err
		}
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, url, &postData)
	if err != nil {
		return
	}

	if agent.HashSigner.KeyExist() {
		hashSignature, err := agent.HashSigner.GenerateSignature(postData.Bytes())
		if err != nil {
			logger.GetLogger().Panicf("failed to generate hash signature: %v", err)
		}
		req.Header.Set(common.HashHeaderKey, hashSignature)
	}

	req.Header.Set("Content-Encoding", compression)
	req.Header.Set("Content-Type", "application/json")
	resp, err = agent.Client.Do(req)
	return
}

func (agent *Agent) SendMetric(key string) error {
	metric, ok := agent.Metrics[key]
	if !ok {
		return errs.ErrorMetricDoesNotExist
	}

	data, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	pathParams := []string{"update"}
	postPath, err := url.JoinPath(agent.address, pathParams...)
	if err != nil {
		return errs.ErrorWrongPath
	}

	if resp, err := agent.Post(postPath, data, true); err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return err
	}
	return nil
}

func (agent *Agent) SendMetricsBatch(inCh <-chan []model.Metric, resultCh chan<- Result) error {
	// defer close(resultCh)

	metrics := <-inCh
	// metrics := agent.toList()

	data, err := json.Marshal(metrics)
	if err != nil {
		resultCh <- Result{err: fmt.Errorf("failed to marshal metrics: %w", err)}
		return err
	}
	logger.GetLogger().Infof("Sending batch metrics count=%d size=%d\n", len(metrics), len(data))

	postPath := agent.address + "/updates/"
	resp, err := agent.Post(postPath, data, true)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		resultCh <- Result{err: fmt.Errorf("failed to send metrics: %w", err)}
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// logger.GetLogger().Errorf("failed to read response body: %v", err)
		resultCh <- Result{err: fmt.Errorf("failed to read response body: %w", err)}
	}
	defer resp.Body.Close()

	logger.GetLogger().Infof("Batch metrics response: status_code=%d, body=%s\n", resp.StatusCode, body)
	return nil
}

func (agent *Agent) toList() (metrics []model.Metric) {
	for _, metric := range agent.Metrics {
		metrics = append(metrics, metric)
	}
	return
}

func (agent *Agent) WorkerPoolCreation() error {
	var wg sync.WaitGroup
	logger.GetLogger().Infof("Starting worker pool creation")

	metrics := agent.toList()

	resultCh := make(chan Result, agent.RateLimit)
	inCh := make(chan []model.Metric, agent.RateLimit)

	for w := range agent.RateLimit {
		wg.Add(1)
		logger.GetLogger().Infof("Worker %d starting", w)
		go func() {
			_ = agent.SendMetricsBatch(inCh, resultCh)
			wg.Done()
		}()
	}

	chunkSize := int(math.Ceil(float64(len(metrics)) / float64(agent.RateLimit)))
	for i := 0; i < len(metrics); i += chunkSize {
		end := i + chunkSize
		if end > len(metrics) {
			end = len(metrics)
		}
		inCh <- metrics[i:end]
	}
	close(inCh)

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for res := range resultCh {
		if res.err != nil {
			logger.GetLogger().Errorf("failed to send metrics batch: %v", res.err)
		}
	}
	return nil
}

func (agent *Agent) SendData(reportInterval int) {
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := agent.WorkerPoolCreation(); err != nil {
			logger.GetLogger().Error(err)
		}
	}
}

func (agent *Agent) Run(pollInterval int, reportInterval int) {
	go agent.Update(pollInterval)
	agent.SendData(reportInterval)
}
