package hashsign

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/dmitastr/yp_observability_service/internal/logger"
)

type RespHashWriter struct {
	http.ResponseWriter
	hg         *HashGenerator
	Body       *bytes.Buffer
	StatusCode *int
}

func NewRespWriter(res http.ResponseWriter, hg *HashGenerator, body *bytes.Buffer) *RespHashWriter {
	return &RespHashWriter{ResponseWriter: res, hg: hg, Body: body}
}

func (c *RespHashWriter) Write(p []byte) (int, error) {
	return c.Body.Write(p)
}

func (c *RespHashWriter) WriteHeader(statusCode int) {
	c.StatusCode = &statusCode
}

type HashGenerator struct {
	Key []byte
}

func NewHashGenerator(key *string) *HashGenerator {
	h := &HashGenerator{}
	if key != nil {
		h.Key = []byte(*key)
	}
	return h
}

func (hg *HashGenerator) Generate(body []byte) []byte {
	h := hmac.New(sha256.New, hg.Key)
	h.Write(body)
	signed := h.Sum(nil)
	return signed
}

func (hg *HashGenerator) IsEqual(body, signed []byte) bool {
	return hmac.Equal(signed, body)
}

func (hg *HashGenerator) KeyExist() bool {
	return len(hg.Key) > 0
}

func (hg *HashGenerator) Decode(hash string) ([]byte, error) {
	return hex.DecodeString(hash)
}

func (hg *HashGenerator) Encode(hash []byte) string {
	return hex.EncodeToString(hash)
}

func (hg *HashGenerator) ReadBody(req *http.Request) ([]byte, error) {
	var body []byte
	if _, err := req.Body.Read(body); err != nil {
		return body, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.GetLogger().Errorf("Error closing body: %v", err)
		}
	}(req.Body)
	return body, nil
}

func (hg *HashGenerator) Verify(hash string, req *http.Request) (bool, error) {
	body, err := hg.ReadBody(req)
	if err != nil {
		return false, err
	}
	hashExpected := hg.Generate(body)
	hashActual, err := hg.Decode(hash)
	if err != nil {
		return false, err
	}
	return hg.IsEqual(hashExpected, hashActual), nil
}

func (hg *HashGenerator) CheckHash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		buf := bytes.NewBuffer(nil)
		useWriter := NewRespWriter(res, hg, buf)
		if hg.KeyExist() {
			hashRequest := req.Header.Get("HashSHA256")
			ok, err := hg.Verify(hashRequest, req)
			if err != nil {
				logger.GetLogger().Errorf("error while calculating hash: %v", err)
				http.Error(res, "error while calculating hash", http.StatusInternalServerError)
				return
			}
			if !ok {
				logger.GetLogger().Error("hashes are not equal")
				http.Error(res, "Error verifying hash", http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(useWriter, req)

		body := useWriter.Body.Bytes()
		if hg.KeyExist() {
			hash := hg.Generate(body)
			hashString := hg.Encode(hash)
			res.Header().Set("HashSHA256", hashString)
		}

		statusCode := http.StatusOK
		if useWriter.StatusCode != nil {
			statusCode = *useWriter.StatusCode
		}
		res.WriteHeader(statusCode)

		_, err := res.Write(body)
		if err != nil {
			logger.GetLogger().Errorf("error while writing response: %v", err)
			return
		}
	})
}
