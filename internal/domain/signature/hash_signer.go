package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// HashSigner can encode and decode message with specific key
type HashSigner struct {
	Key *string
}

func NewHashSigner(key *string) *HashSigner {
	hs := &HashSigner{}
	if key != nil && *key != "" {
		hs.Key = key
	}
	return hs
}

func (hs *HashSigner) KeyExist() bool {
	return hs.Key != nil
}

func (hs *HashSigner) Decode(src string) ([]byte, error) {
	return hex.DecodeString(src)
}

func (hs *HashSigner) Encode(src []byte) string {
	return hex.EncodeToString(src)
}

// GenerateSignature signs bytes data with key using sha256 method
func (hs *HashSigner) GenerateSignature(body []byte) (string, error) {
	if !hs.KeyExist() {
		return "", errors.New("key does not exist")
	}

	h := hmac.New(sha256.New, []byte(*hs.Key))
	_, err := h.Write(body)
	if err != nil {
		return "", fmt.Errorf("failed to write to hash: %w", err)
	}

	signed := h.Sum(nil)
	return hs.Encode(signed), nil
}

func (hs *HashSigner) Verify(signatureActual, signatureExpected string) bool {
	return signatureActual == signatureExpected
}
