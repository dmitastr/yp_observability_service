package data

import (
	"encoding/json"
	"time"

	"github.com/dmitastr/yp_observability_service/internal/model"
)

type Data struct {
	MetricNames []string `json:"metrics"`
	IP          string   `json:"ip_address"`
	Timestamp   int64    `json:"ts"`
}

func NewData(metrics []models.Metrics, ipAddress string) *Data {
	var names []string
	for _, m := range metrics {
		names = append(names, m.ID)
	}

	return &Data{
		MetricNames: names,
		IP:          ipAddress,
		Timestamp:   time.Now().Unix(),
	}
}

func (data Data) Marshal() ([]byte, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return payload, nil
}
