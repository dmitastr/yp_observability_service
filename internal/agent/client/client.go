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

	"github.com/dmitastr/yp_observability_service/internal/agent/rsaencoder"
	"github.com/dmitastr/yp_observability_service/internal/domain/signature"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	model "github.com/dmitastr/yp_observability_service/internal/agent/metric"
	"github.com/dmitastr/yp_observability_service/internal/common"
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

// Run initialized [cobra.Command] for args parsing and starts the server
func Run() error {
	rootCmd := &cobra.Command{
		Use: "YP observability agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := logger.Initialize(); err != nil {
				return err
			}

			if cfgPath := viper.GetString("config"); cfgPath != "" {
				viper.SetConfigFile(cfgPath)
				if err := viper.ReadInConfig(); err != nil {
					logger.Errorf("Error reading config file, %s\n", err)
				}
			}

			var cfg config.Config
			// Unmarshal the configuration into the Config struct
			if err := viper.Unmarshal(&cfg); err != nil {
				logger.Errorf("Unable to decode into struct, %v\n", err)
				return err
			}
			agent := NewAgent(cfg)

			logger.Infof("Starting client for server=%s, poll interval=%d, report interval=%d",
				*cfg.Address,
				*cfg.PollInterval,
				*cfg.ReportInterval,
			)

			agent.Run(*cfg.PollInterval, *cfg.ReportInterval)

			return nil
		},
	}

	rootCmd.Flags().StringP("address", "a", "localhost:8080", "set server host and port")
	rootCmd.Flags().IntP("report_interval", "r", 10, "frequency of data sending to server in seconds")
	rootCmd.Flags().IntP("poll_interval", "p", 10, "frequency of metric polling from source in seconds")
	rootCmd.Flags().IntP("rate_limit", "l", 3, "rate limit")
	rootCmd.Flags().String("k", "", "key for request signing")
	rootCmd.Flags().String("crypto-key", "", "path to file with public key")
	rootCmd.Flags().StringP("config", "c", "", "path to config file")

	_ = viper.BindPFlags(rootCmd.Flags())

	viper.AutomaticEnv()

	// Bind environment variables
	_ = viper.BindEnv("address", "ADDRESS")
	_ = viper.BindEnv("k", "KEY")
	_ = viper.BindEnv("report_interval", "REPORT_INTERVAL")
	_ = viper.BindEnv("poll_interval", "POLL_INTERVAL")
	_ = viper.BindEnv("rate_limit", "RATE_LIMIT")
	_ = viper.BindEnv("crypto-key", "CRYPTO_KEY")
	_ = viper.BindEnv("config", "CONFIG")

	return rootCmd.Execute()

}

type Result struct {
	err error
}

type Agent struct {
	sync.Mutex
	Metrics    map[string]model.Metric
	Client     *retryablehttp.Client
	address    string
	HashSigner *signature.HashSigner
	RateLimit  int
	encoder    *rsaencoder.Encoder
}

func NewAgent(cfg config.Config) *Agent {
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

	if cfg.PublicKeyFile != nil && *cfg.PublicKeyFile != "" {
		agent.encoder = rsaencoder.NewEncoder(*cfg.PublicKeyFile)
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

func (agent *Agent) UpdateSysUtilMetrics() {
	agent.Mutex.Lock()
	defer agent.Mutex.Unlock()

	v, _ := mem.VirtualMemory()

	agent.UpdateMetricValueGauge(TotalMemory, float64(v.Total))
	agent.UpdateMetricValueGauge(FreeMemory, float64(v.Free))
	cpuStats, _ := cpu.Info()
	if len(cpuStats) > 0 {
		agent.UpdateMetricValueGauge(CPUutilization1, float64(cpuStats[0].CPU))
	}
}

func (agent *Agent) UpdateRuntimeMetrics() {
	agent.Mutex.Lock()
	defer agent.Mutex.Unlock()

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
}

func (agent *Agent) Update(pollInterval int) {

	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		go agent.UpdateRuntimeMetrics()
		go agent.UpdateSysUtilMetrics()
	}
}

func (agent *Agent) Post(url string, data []byte, compressed bool) (resp *http.Response, err error) {
	var postData bytes.Buffer
	var compression string

	if compressed {
		data, err = agent.compress(data)
		if err != nil {
			logger.Errorf("failed to close gzip writer: %v", err)
		}
		compression = "gzip"
	}

	encoded, err := agent.Encode(data)
	if err != nil {
		logger.Errorf("failed to encode post data: %v", err)
	} else {
		data = encoded
	}

	if _, err = postData.Write(data); err != nil {
		logger.Errorf("failed to post data: %v", err)
		return
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, url, &postData)
	if err != nil {
		return
	}

	if agent.HashSigner.KeyExist() {
		hashSignature, err := agent.HashSigner.GenerateSignature(postData.Bytes())
		if err != nil {
			logger.Panicf("failed to generate hash signature: %v", err)
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
	logger.Infof("Sending batch metrics count=%d size=%d\n", len(metrics), len(data))

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
		resultCh <- Result{err: fmt.Errorf("failed to read response body: %w", err)}
	}
	defer resp.Body.Close()

	logger.Infof("Batch metrics response: status_code=%d, body=%s\n", resp.StatusCode, body)
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
	logger.Infof("Starting worker pool creation")

	metrics := agent.toList()

	resultCh := make(chan Result, agent.RateLimit)
	inCh := make(chan []model.Metric, agent.RateLimit)

	for w := range agent.RateLimit {
		wg.Add(1)
		logger.Infof("Worker %d starting", w)
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
			logger.Errorf("failed to send metrics batch: %v", res.err)
		}
	}
	return nil
}

func (agent *Agent) SendData(reportInterval int) {
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := agent.WorkerPoolCreation(); err != nil {
			logger.Error(err)
		}
	}
}

func (agent *Agent) Run(pollInterval int, reportInterval int) {
	go agent.Update(pollInterval)
	agent.SendData(reportInterval)
}

func (agent *Agent) Encode(data []byte) ([]byte, error) {
	if agent.encoder != nil {
		return agent.encoder.Encode(data)
	}
	return data, nil
}

func (agent *Agent) compress(data []byte) ([]byte, error) {
	var compressed bytes.Buffer

	gw := gzip.NewWriter(&compressed)
	if _, err := gw.Write(data); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		logger.Errorf("failed to close gzip writer: %v", err)
	}
	return compressed.Bytes(), nil
}
