package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type Writer struct {
	rw http.ResponseWriter
	gz *gzip.Writer
}

func NewCompressWriter(res http.ResponseWriter) *Writer {
	return &Writer{rw: res, gz: gzip.NewWriter(res)}
}

func (c *Writer) Write(p []byte) (int, error) {
	return c.gz.Write(p)
}

func (c *Writer) Header() http.Header {
	return c.rw.Header()
}

func (c *Writer) WriteHeader(statusCode int) {
	if statusCode < 600 {
		c.rw.Header().Set("Content-Encoding", "gzip")
	}
	c.rw.WriteHeader(statusCode)
}

func (c *Writer) Close() error {
	return c.gz.Close()
}

type Reader struct {
	r  io.ReadCloser
	gz *gzip.Reader
}

func NewCompressReader(reader io.ReadCloser) (c *Reader, err error) {
	gz, err := gzip.NewReader(reader)
	if err != nil {
		return
	}
	c = &Reader{r: reader, gz: gz}
	return
}

func (c *Reader) Read(p []byte) (int, error) {
	return c.gz.Read(p)
}

func (c *Reader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.gz.Close()
}

func Handler(next http.Handler) http.Handler {
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
			defer func(cw *Writer) {
				_ = cw.Close()
			}(cw)
		}

		next.ServeHTTP(useWriter, req)
	})
}
