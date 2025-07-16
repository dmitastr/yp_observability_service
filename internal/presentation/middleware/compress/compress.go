package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	rw http.ResponseWriter
	gz *gzip.Writer
}

func NewCompressWriter(res http.ResponseWriter) *compressWriter {
	return &compressWriter{rw: res, gz: gzip.NewWriter(res)}
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.gz.Write(p)
}

func (c *compressWriter) Header() http.Header {
	return c.rw.Header()
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.rw.Header().Set("Content-Encoding", "gzip")
	c.rw.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.gz.Close()
}

type compressReader struct {
    r  io.ReadCloser
    gz *gzip.Reader
}

func NewCompressReader(reader io.ReadCloser) (c *compressReader, err error) {
	gz, err := gzip.NewReader(reader)
	if err != nil {
		return
	}
	c = &compressReader{r: reader, gz: gz}
	return
}

func (c *compressReader) Read(p []byte) (int, error) {
	return c.gz.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.gz.Close()
}

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
