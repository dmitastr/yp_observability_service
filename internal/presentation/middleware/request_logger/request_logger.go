package requestlogger

import (
	"net/http"
	"time"

	"github.com/dmitastr/yp_observability_service/internal/logger"
)

// ResponseData save response stats
type ResponseData struct {
	BodySize   int
	StatusCode int
}

// LoggingResponseWriter implements [http.ResponseWriter] interface and is used to log request and response
type LoggingResponseWriter struct {
	w            http.ResponseWriter
	ResponseData *ResponseData
}

func NewLoggingResponseWriter(w http.ResponseWriter) LoggingResponseWriter {
	return LoggingResponseWriter{
		w:            w,
		ResponseData: &ResponseData{},
	}
}

// Write calculate body size and saves it to [ResponseData]
func (rww LoggingResponseWriter) Write(buf []byte) (int, error) {
	size, err := rww.w.Write(buf)
	rww.ResponseData.BodySize += size
	return size, err
}

// Header gets [http.Header] to implement [http.ResponseWriter] interface
func (rww LoggingResponseWriter) Header() http.Header {
	return rww.w.Header()
}

// WriteHeader saves status code to [ResponseData] and writes it to original [http.ResponseWriter]
func (rww LoggingResponseWriter) WriteHeader(statusCode int) {
	rww.ResponseData.StatusCode = statusCode
	rww.w.WriteHeader(statusCode)
}

// RequestLogger middleware logs request method, URI and response status code, calculates time of execution
// and response body size
func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rww := NewLoggingResponseWriter(w)
		logger.Infof("receive request: method=%s, URI=%s", r.Method, r.RequestURI)

		h.ServeHTTP(rww, r)

		execTime := time.Since(start)
		logger.Infof("sending response: execution time=%s, status code=%d, body size=%d",
			execTime.String(),
			rww.ResponseData.StatusCode,
			rww.ResponseData.BodySize,
		)
	})
}
