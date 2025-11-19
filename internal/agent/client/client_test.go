package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	agentenvconfig "github.com/dmitastr/yp_observability_service/internal/config/env_parser/agent/agent_env_config"
	"github.com/stretchr/testify/assert"
)

// DONE
func TestAgent_UpdateMetricValueCounter(t *testing.T) {
	addr := `localhost:8080`
	cfg := agentenvconfig.New(addr, 0, 0, "", 1)
	agent, _ := NewAgent(cfg)
	type args struct {
		key   string
		value int64
	}
	tests := []struct {
		name      string
		params    []args
		wantValue any
		wantKey   string
	}{
		{
			name:      "Valid input",
			params:    []args{{key: "abc1", value: 10}},
			wantKey:   "abc1",
			wantValue: int64(10),
		},
		{
			name:      "Update value",
			params:    []args{{key: "abc", value: 10}, {key: "abc", value: 20}},
			wantKey:   "abc",
			wantValue: int64(30),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, param := range tt.params {
				agent.UpdateMetricValueCounter(param.key, param.value)
			}
			actual := agent.Metrics[tt.wantKey].GetValue()
			assert.Equal(t, tt.wantValue, actual)
		})
	}
}

func TestAgent_UpdateMetricValueGauge(t *testing.T) {
	addr := `localhost:8080`
	cfg := agentenvconfig.New(addr, 0, 0, "", 1)
	agent, _ := NewAgent(cfg)

	type args struct {
		key   string
		value float64
	}
	tests := []struct {
		name      string
		params    []args
		wantValue any
		wantKey   string
	}{
		{
			name:      "Valid input",
			params:    []args{{key: "abc", value: 10}},
			wantKey:   "abc",
			wantValue: float64(10.0),
		},
		{
			name:      "Update value",
			params:    []args{{key: "abc", value: 10}, {key: "abc", value: 20}},
			wantKey:   "abc",
			wantValue: float64(20.0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, param := range tt.params {
				agent.UpdateMetricValueGauge(param.key, param.value)
			}
			actual := agent.Metrics[tt.wantKey].GetValue()
			assert.Equal(t, tt.wantValue, actual)
		})
	}
}

func TestAgent_SendMetric(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "send request to url=%s\n", r.URL)
	}))
	defer srv.Close()

	type args struct {
		key   string
		value int64
	}
	tests := []struct {
		name      string
		metrics   args
		keyToSend string
		wantErr   bool
	}{
		{
			name:      "valid input",
			metrics:   args{"abc", 10},
			keyToSend: "abc",
			wantErr:   false,
		},
		{
			name:      "missing metric key",
			metrics:   args{"abc", 10},
			keyToSend: "sdf",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := agentenvconfig.New(srv.URL, 0, 0, "", 1)
			agent, _ := NewAgent(cfg)
			agent.UpdateMetricValueCounter("abc", 1)
			err := agent.SendMetric(t.Context(), tt.keyToSend)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
