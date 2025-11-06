package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/dmitastr/yp_observability_service/internal/logger"
)

// CompressWriter implements [http.ResponseWriter] interface and is used to compress response
type CompressWriter struct {
	rw   http.ResponseWriter
	gz   *gzip.Writer
	code int
}

func NewCompressWriter(res http.ResponseWriter) *CompressWriter {
	return &CompressWriter{rw: res, gz: gzip.NewWriter(res)}
}

// Write compress the data and writes it to the original [http.ResponseWriter]
func (c *CompressWriter) Write(p []byte) (int, error) {
	c.rw.Header().Set("Content-Encoding", "gzip")
	return c.gz.Write(p)
}

// Header gets [http.Header] to implement [http.ResponseWriter] interface
func (c *CompressWriter) Header() http.Header {
	return c.rw.Header()
}

// WriteHeader checks status code and adds Content-encoding header to the response
func (c *CompressWriter) WriteHeader(statusCode int) {
	c.rw.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	return c.gz.Close()
}

// CompressReader used for decompressing request
type CompressReader struct {
	r  io.ReadCloser
	gz *gzip.Reader
}

func NewCompressReader(reader io.ReadCloser) (c *CompressReader, err error) {
	gz, err := gzip.NewReader(reader)
	if err != nil {
		logger.Errorf("error creating gzip reader: %v", err)
		return
	}
	c = &CompressReader{r: reader, gz: gz}
	return
}

// Read decompress the data and writes it to p
func (c *CompressReader) Read(p []byte) (int, error) {
	return c.gz.Read(p)
}

// Close closes the reader to avoid memory leakage
func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.gz.Close()
}

// SetReader checks if request was compressed and set reader for decompression
func (c *CompressReader) SetReader(req *http.Request) {
	contentEncoding := req.Header.Get("Content-Encoding")
	if strings.Contains(contentEncoding, "gzip") {
		req.Body = c
	}
}

// HandleCompression is a middleware that handles compression and decompression if appropriate headers are set
func HandleCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		useWriter := res
		contentEncoding := req.Header.Get("Content-Encoding")
		if strings.Contains(contentEncoding, "gzip") {
			cr, err := NewCompressReader(req.Body)
			if err != nil {
				logger.Errorf("error creating compress reader: %v", err)
				http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			req.Body = cr
		}

		acceptEncoding := req.Header.Get("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") {
			cw := NewCompressWriter(res)
			useWriter = cw
			defer cw.Close()
		}

		next.ServeHTTP(useWriter, req)
	})
}

// HandleDecompression checks if request was compressed and set reader for decompression
func HandleDecompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		contentEncoding := req.Header.Get("Content-Encoding")
		if strings.Contains(contentEncoding, "gzip") {
			cr, err := NewCompressReader(req.Body)
			if err != nil {
				logger.Errorf("error creating compress reader: %v", err)
				http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			req.Body = cr
		}

		next.ServeHTTP(res, req)
	})
}
