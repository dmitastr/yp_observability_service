package rsaencoder

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

type Encoder struct {
	pubKey *rsa.PublicKey
}

func NewEncoder(publicKeyFile string) (*Encoder, error) {
	c := new(Encoder)
	key, err := os.ReadFile(publicKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}
	publicPem, _ := pem.Decode(key)
	if publicPem == nil {
		return nil, fmt.Errorf("public key not found")
	}
	publicKey, err := x509.ParseCertificate(publicPem.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	pubKeyParsed, ok := publicKey.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to parse public key")
	}
	c.pubKey = pubKeyParsed
	return c, nil
}

func (c *Encoder) Encode(data []byte) ([]byte, error) {
	if c.pubKey == nil {
		return data, errors.New("public key is nil")
	}
	return rsa.EncryptPKCS1v15(rand.Reader, c.pubKey, data)
}
