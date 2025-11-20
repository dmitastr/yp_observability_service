package compress

import (
	"bytes"
	"compress/gzip"

	"github.com/dmitastr/yp_observability_service/internal/logger"
)

type Compressor struct {
}

func NewCompressor() *Compressor {
	return &Compressor{}
}

func (c *Compressor) Compress(data []byte) ([]byte, error) {
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
