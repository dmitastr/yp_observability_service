package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

const bodySizeForCompression int = 100

// CompressWriter implements [http.ResponseWriter] interface and is used to compress response
type CompressWriter struct {
	rw http.ResponseWriter
	gz *gzip.Writer
}

func NewCompressWriter(res http.ResponseWriter) *CompressWriter {
	return &CompressWriter{rw: res, gz: gzip.NewWriter(res)}
}

// Write compress the data and writes it to the original [http.ResponseWriter]
func (c *CompressWriter) Write(p []byte) (int, error) {
	if len(p) > bodySizeForCompression {
		return c.gz.Write(p)
	}
	return c.rw.Write(p)
}

// Header gets [http.Header] to implement [http.ResponseWriter] interface
func (c *CompressWriter) Header() http.Header {
	return c.rw.Header()
}

// WriteHeader checks status code and adds Content-encoding header to the response
func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < 600 {
		c.rw.Header().Set("Content-Encoding", "gzip")
	}
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

// CompressMiddleware is a middleware that handles compression and decompression if appropriate headers are set
func CompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		useWriter := res
		contentEncoding := req.Header.Get("Content-Encoding")
		if strings.Contains(contentEncoding, "gzip") {
			cr, err := NewCompressReader(req.Body)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
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
