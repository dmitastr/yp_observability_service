package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dmitastr/yp_observability_service/internal/agent/compress"
	"github.com/dmitastr/yp_observability_service/internal/agent/models"
	"github.com/dmitastr/yp_observability_service/internal/agent/rsaencoder"
	"github.com/dmitastr/yp_observability_service/internal/common"
	config "github.com/dmitastr/yp_observability_service/internal/config/env_parser/agent/agent_env_config"
	"github.com/dmitastr/yp_observability_service/internal/domain/signature"
	"github.com/dmitastr/yp_observability_service/internal/errs"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/net/context"
)

type Client struct {
	client     *retryablehttp.Client
	address    string
	encoder    *rsaencoder.Encoder
	compressor *compress.Compressor
	hashSigner *signature.HashSigner
}

func NewClient(cfg config.Config) (*Client, error) {
	httpClient := retryablehttp.NewClient()
	httpClient.HTTPClient.Timeout = time.Millisecond * 300
	httpClient.RetryMax = 3
	httpClient.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		return time.Second * time.Duration(2*attemptNum+1)
	}

	address := *cfg.Address
	if !strings.Contains(address, "http") {
		address = "http://" + address
	}

	client := &Client{
		client:     httpClient,
		address:    address,
		compressor: compress.NewCompressor(),
		hashSigner: signature.NewHashSigner(cfg.Key),
	}

	if cfg.PublicKeyFile != nil && *cfg.PublicKeyFile != "" {
		encoder, err := rsaencoder.NewEncoder(*cfg.PublicKeyFile)
		if err != nil {
			return nil, fmt.Errorf("error creating rsa encoder: %w", err)
		}
		client.encoder = encoder
	}

	return client, nil
}

func (c *Client) Encode(data []byte) ([]byte, error) {
	if c.encoder != nil {
		return c.encoder.Encode(data)
	}
	return data, nil
}

func (c *Client) Post(ctx context.Context, url string, data []byte, compressed bool) (resp *http.Response, err error) {
	var postData bytes.Buffer
	var compression string

	if compressed {
		data, err = c.compressor.Compress(data)
		if err != nil {
			logger.Errorf("failed to close gzip writer: %v", err)
		}
		compression = "gzip"
	}

	encoded, err := c.Encode(data)
	if err != nil {
		logger.Errorf("failed to encode post data: %v", err)
	} else {
		data = encoded
	}

	if _, err = postData.Write(data); err != nil {
		logger.Errorf("failed to post data: %v", err)
		return
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, url, &postData)
	if err != nil {
		return
	}

	ipValue := ctx.Value(models.RealIP{})

	var ip string
	ip, ok := ipValue.(string)
	if !ok {
		ip = "127.0.0.1"
	}
	req.Header.Set("X-Real-IP", ip)

	if c.hashSigner.KeyExist() {
		hashSignature, err := c.hashSigner.GenerateSignature(postData.Bytes())
		if err != nil {
			logger.Panicf("failed to generate hash signature: %v", err)
		}
		req.Header.Set(common.HashHeaderKey, hashSignature)
	}

	req.Header.Set("Content-Encoding", compression)
	req.Header.Set("Content-Type", "application/json")
	resp, err = c.client.Do(req)
	return
}

func (c *Client) SendMetric(ctx context.Context, metric models.Metric) error {
	data, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	pathParams := []string{"update"}
	postPath, err := url.JoinPath(c.address, pathParams...)
	if err != nil {
		return errs.ErrorWrongPath
	}

	if resp, err := c.Post(ctx, postPath, data, true); err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return err
	}
	return nil
}

func (c *Client) SendMetricsBatch(ctx context.Context, metrics []models.Metric) error {
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}
	logger.Infof("Sending batch metrics count=%d size=%d\n", len(metrics), len(data))

	postPath := c.address + "/updates/"
	resp, err := c.Post(ctx, postPath, data, true)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return fmt.Errorf("failed to send metrics: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	defer resp.Body.Close()

	logger.Infof("Batch metrics response: status_code=%d, body=%s\n", resp.StatusCode, body)
	return nil
}

func (c *Client) Close(ctx context.Context) error {
	return nil
}
