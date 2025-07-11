package requestlogger

import (
	"net/http"
	"time"

	"github.com/dmitastr/yp_observability_service/internal/logger"
)

type ResponseData struct {
	BodySize   int
	StatusCode int
}

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

func (rww LoggingResponseWriter) Write(buf []byte) (int, error) {
	size, err := rww.w.Write(buf)
	rww.ResponseData.BodySize += size
	return size, err
}

func (rww LoggingResponseWriter) Header() http.Header {
	return rww.w.Header()
}

func (rww LoggingResponseWriter) WriteHeader(statusCode int) {
	rww.w.WriteHeader(statusCode)
	rww.ResponseData.StatusCode = statusCode
}


// Сведения о запросах должны содержать URI, метод запроса и время, затраченное на его выполнение.
// Сведения об ответах должны содержать код статуса и размер содержимого ответа.
func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rww := NewLoggingResponseWriter(w)
		logger.GetLogger().Infof("receive request: method=%s, URI=%s", r.Method, r.RequestURI)

		h.ServeHTTP(rww, r)

		execTime := time.Since(start)
		logger.GetLogger().Infof("sending response: execution time=%s, status code=%d, body size=%d", 
			execTime.String(), 
			rww.ResponseData.StatusCode,
			rww.ResponseData.BodySize,
		)
	})
}
