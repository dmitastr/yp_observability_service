package certdecode

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"os"

	"github.com/dmitastr/yp_observability_service/internal/logger"
)

// CertDecoder is a middleware for decoding messages using private key
type CertDecoder struct {
	privateKey *rsa.PrivateKey
}

func NewCertDecoder(privateFile string) *CertDecoder {
	c := new(CertDecoder)
	if privateFile == "" {
		return c
	}

	privateKeyBytes, err := os.ReadFile(privateFile)
	if err != nil {
		logger.Infof("Failed to read private key file: %v\n", err)
		return c
	}
	privateKeyBlock, _ := pem.Decode(privateKeyBytes)
	if privateKeyBlock == nil {
		logger.Info("Private key not found")
		return c
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		logger.Infof("Failed to parse private key: %v\n", err)
		return c
	}
	c.privateKey = privateKey
	return c
}

// Decode decodes incoming message using private key
func (c *CertDecoder) Decode(message []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, c.privateKey, message)
}

// Handle checks if private key is set and decodes incoming message
func (c *CertDecoder) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.privateKey != nil {
			body, err := io.ReadAll(r.Body)
			defer r.Body.Close()

			if err != nil {
				logger.Infof("Failed to read request body: %v\n", err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			bodyDecoded, err := c.Decode(body)
			if err != nil {
				logger.Infof("Failed to decode body: %v\n", err)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyDecoded))
		}
		next.ServeHTTP(w, r)

	})
}
