package agent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/nickzhog/devops-tool/internal/agent/config"
	"github.com/nickzhog/devops-tool/pkg/logging"
	"github.com/nickzhog/devops-tool/pkg/metric"
	"github.com/stretchr/testify/assert"
)

func Test_agent_SendMetricsHTTP(t *testing.T) {
	cfg := &config.Config{}
	logger := logging.GetLogger()
	cfg.Settings.Address = "http://localhost"

	tests := []struct {
		name        string
		metricsJSON string
	}{
		{
			name:        "case #1",
			metricsJSON: `[{"id":"test1","type":"counter","delta":12}]`,
		},
		{
			name:        "case #2",
			metricsJSON: `[{"id":"test1","type":"counter","delta":12}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			assert := assert.New(t)

			httpmock.RegisterResponder(http.MethodPost, "http://localhost/update",
				func(req *http.Request) (*http.Response, error) {
					body, err := io.ReadAll(req.Body)
					if err != nil {
						assert.Fail(err.Error())
						return httpmock.NewStringResponse(http.StatusBadRequest, ""), err
					}

					var m metric.Metric
					err = json.Unmarshal(body, &m)
					if err != nil {
						assert.Fail(err.Error())
						return httpmock.NewStringResponse(http.StatusBadRequest, ""), err
					}

					return nil, nil
				})

			agentStorage := NewAgent(cfg, logger)

			var metrics []metric.Metric
			err := json.Unmarshal([]byte(tt.metricsJSON), &metrics)
			assert.NoError(err)

			err = agentStorage.ImportMetrics(metrics)
			assert.NoError(err)

			agentStorage.SendMetricsHTTP(context.Background())

		})
	}
}

func Test_agent_ImportMetrics(t *testing.T) {
	cfg := &config.Config{}
	logger := logging.GetLogger()

	tests := []struct {
		name        string
		metricsJSON string
	}{
		{
			name:        "case #1",
			metricsJSON: `[{"id":"test1","type":"gauge","value":1222}]`,
		},
		{
			name:        "case #2",
			metricsJSON: `[{"id":"test2","type":"counter","delta":90}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAgent(cfg, logger)
			assert := assert.New(t)

			var metrics []metric.Metric
			err := json.Unmarshal([]byte(tt.metricsJSON), &metrics)
			assert.NoError(err)

			err = a.ImportMetrics(metrics)
			assert.NoError(err)

			metrics = a.ExportMetrics()

			jsonData, err := json.Marshal(metrics)
			assert.NoError(err)

			assert.JSONEq(tt.metricsJSON, string(jsonData))
		})
	}
}
