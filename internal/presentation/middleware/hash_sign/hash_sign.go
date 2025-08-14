package hash_sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"io"
	"net/http"

	"github.com/dmitastr/yp_observability_service/internal/logger"
)

type HashGenerator struct {
	Key []byte
}

func NewHashGenerator(key string) *HashGenerator {
	h := &HashGenerator{}
	if key != "" {
		h.Key = []byte(key)
	}
	return h
}

func (hg *HashGenerator) Generate(body []byte) []byte {
	h := hmac.New(sha256.New, hg.Key)
	h.Write(body)
	signed := h.Sum(nil)
	return signed
}

func (hg *HashGenerator) Verify(body, signed []byte) bool {
	return hmac.Equal(signed, body)
}

func (hg *HashGenerator) KeyExist() bool {
	return hg.Key != nil
}

func (hg *HashGenerator) CheckHash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		hashRequest := req.Header.Get("HashSHA256")
		if hashRequest != "" && hg.KeyExist() {
			var body []byte
			if _, err := req.Body.Read(body); err != nil {
				logger.GetLogger().Errorf("Error reading body: %v", err)
				http.Error(res, "Error reading body", http.StatusBadRequest)
				return
			}
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					logger.GetLogger().Errorf("Error closing body: %v", err)
				}
			}(req.Body)

			hashExpected := hg.Generate(body)
			if !hg.Verify(hashExpected, []byte(hashRequest)) {
				logger.GetLogger().Errorf("hashes are not equal")
				http.Error(res, "Error verifying hash", http.StatusBadRequest)
				return
			}

			logger.GetLogger().Infof("Receive signed data with hash=%s", hashRequest)
		}

		next.ServeHTTP(res, req)
	})
}
