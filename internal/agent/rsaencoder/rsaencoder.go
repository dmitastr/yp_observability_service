package rsaencoder

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

	"github.com/dmitastr/yp_observability_service/internal/logger"
)

type Encoder struct {
	pubKey *rsa.PublicKey
}

func NewEncoder(publicKeyFile string) (c *Encoder) {
	c = new(Encoder)
	key, err := os.ReadFile(publicKeyFile)
	if err != nil {
		logger.Infof("Failed to read private key file: %v\n", err)
		return
	}
	publicPem, _ := pem.Decode(key)
	if publicPem == nil {
		logger.Info("Public key not found")
		return
	}
	publicKey, err := x509.ParseCertificate(publicPem.Bytes)
	if err != nil {
		logger.Infof("Failed to parse public key: %v\n", err)
		return
	}
	pubKeyParsed, ok := publicKey.PublicKey.(*rsa.PublicKey)
	if !ok {
		logger.Infof("Failed to parse public key\n")
		return
	}
	c.pubKey = pubKeyParsed
	return
}

func (c *Encoder) Encode(data []byte) ([]byte, error) {
	if c.pubKey == nil {
		return data, errors.New("public key is nil")
	}
	return rsa.EncryptPKCS1v15(rand.Reader, c.pubKey, data)
}
