package hash

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dmitastr/yp_observability_service/internal/common"
	serverenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/server/server_env_config"
	"github.com/dmitastr/yp_observability_service/internal/domain/signature"
	"github.com/dmitastr/yp_observability_service/internal/logger"
)

type writer struct {
	http.ResponseWriter
	Body       *bytes.Buffer
	StatusCode *int
}

func newWriter(w http.ResponseWriter) *writer {
	return &writer{ResponseWriter: w, Body: bytes.NewBuffer([]byte{})}
}

func (w *writer) Write(p []byte) (int, error) {
	w.Body.Write(p)
	return len(p), nil
}

func (w *writer) WriteHeader(statusCode int) {
	w.StatusCode = &statusCode
}

type SignedChecker struct {
	HashSigner *signature.HashSigner
}

func NewSignedChecker(cfg serverenvconfig.Config) *SignedChecker {
	hs := signature.NewHashSigner(cfg.Key)
	return &SignedChecker{HashSigner: hs}
}

// Handle middleware is used for ensuring that request is signed correctly and signs response
func (s *SignedChecker) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		w := newWriter(res)
		keyExist := s.HashSigner.KeyExist()

		hashRequest := req.Header.Get(common.HashHeaderKey)
		if hashRequest != "" {
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				err = fmt.Errorf("error getting body from request: %w", err)
				logger.Error(err)
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}

			hashActual, err := s.HashSigner.GenerateSignature(bodyBytes)
			if err != nil {
				err = fmt.Errorf("error generating hash: %w", err)
				logger.Error(err)
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}

			if !s.HashSigner.Verify(hashRequest, hashActual) {
				err := errors.New("hashes are not equal")
				logger.Error(err)
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			logger.Info("hash signature verified")
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		next.ServeHTTP(w, req)

		respBody := w.Body.Bytes()
		if keyExist {
			signedBody, err := s.HashSigner.GenerateSignature(respBody)
			if err != nil {
				err = fmt.Errorf("error generating hash for response: %w", err)
				logger.Error(err)
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			res.Header().Set(common.HashHeaderKey, signedBody)
		}

		statusCode := http.StatusOK
		if w.StatusCode != nil {
			statusCode = *w.StatusCode
		}
		res.WriteHeader(statusCode)

		_, err := res.Write(respBody)
		if err != nil {
			err = fmt.Errorf("error writing to original response body: %w", err)
			logger.Error(err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
