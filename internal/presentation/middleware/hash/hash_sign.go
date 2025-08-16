package hash

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	serverenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/logger"
	"github.com/dmitastr/yp_observability_service/internal/signature"
)

type SignedChecker struct {
	HashSigner *signature.HashSigner
}

func NewSignedChecker(cfg serverenvconfig.Config) *SignedChecker {
	hs := signature.NewHashSigner(cfg.Key)
	return &SignedChecker{HashSigner: hs}
}

func (s *SignedChecker) Check(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		hashRequest := req.Header.Get("HashSHA256")
		if hashRequest != "" {
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				err = fmt.Errorf("error getting body from request: %w", err)
				logger.GetLogger().Error(err)
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}

			hashActual, err := s.HashSigner.GenerateSignature(bodyBytes)
			if err != nil {
				err = fmt.Errorf("error generating hash: %w", err)
				logger.GetLogger().Error(err)
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}

			if !s.HashSigner.Verify(hashRequest, hashActual) {
				err := errors.New("hashes are not equal")
				logger.GetLogger().Error(err)
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			logger.GetLogger().Info("hash signature verified")
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		next.ServeHTTP(res, req)
	})
}
