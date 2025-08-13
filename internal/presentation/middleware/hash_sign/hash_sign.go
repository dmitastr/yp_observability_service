package hash_sign

import (
	"net/http"

	"github.com/dmitastr/yp_observability_service/internal/logger"
)

func HashSigningMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		hashRequest := req.Header.Get("HashSHA256")
		if hashRequest != "" {
			logger.GetLogger().Infof("Receive signed data with hash=%s", hashRequest)
		}

		next.ServeHTTP(res, req)
	})
}
