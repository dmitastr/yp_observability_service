package listener

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/dmitastr/yp_observability_service/internal/domain/audit/data"
)

type URLListener struct {
	url    string
	client *http.Client
}

func NewURLListener(url string) *URLListener {
	return &URLListener{url: url, client: &http.Client{}}
}

func (l *URLListener) Notify(data *data.Data) error {
	var postData bytes.Buffer

	if err := json.NewEncoder(&postData).Encode(data); err != nil {
		return err
	}

	r, err := l.client.Post(l.url, "application/json", &postData)
	defer func() {
		_ = r.Body.Close()
	}()

	return err
}
