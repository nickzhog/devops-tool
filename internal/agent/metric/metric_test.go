package metric

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/nickzhog/devops-tool/internal/agent/config"
	serverMetric "github.com/nickzhog/devops-tool/internal/server/metric"
	"github.com/nickzhog/devops-tool/internal/server/metric/cache"
	"github.com/nickzhog/devops-tool/pkg/logging"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_SendMetrics(t *testing.T) {
	cfg := &config.Config{}
	logger := logging.GetLogger()
	cfg.Settings.Address = "http://localhost"

	tests := []struct {
		name    string
		metrics string
	}{
		{
			name:    "case #1 / positive",
			metrics: `[{"id":"test1","type":"counter","delta":12},{"id":"test2","type":"gauge","value":12.1}]`,
		},
		{
			name:    "case #2 / positive",
			metrics: `[{"id":"test1","type":"counter","delta":12},{"id":"test1","type":"counter","delta":12}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			mockStorage := cache.NewMemStorage()

			httpmock.RegisterResponder(http.MethodPost, "http://localhost/update",
				func(req *http.Request) (*http.Response, error) {
					body, err := io.ReadAll(req.Body)
					if err != nil {
						return httpmock.NewStringResponse(http.StatusBadRequest, ""), err
					}

					var elem serverMetric.Metric
					err = json.Unmarshal(body, &elem)
					if err != nil {
						return httpmock.NewStringResponse(http.StatusBadRequest, ""), err
					}

					err = mockStorage.UpsertMetric(req.Context(), &elem)
					if err != nil {
						return httpmock.NewStringResponse(http.StatusBadRequest, ""), err
					}

					output := elem.Marshal()
					resp, err := httpmock.NewJsonResponse(200, output)
					if err != nil {
						return httpmock.NewStringResponse(http.StatusBadGateway, ""), err
					}
					return resp, nil
				})

			//////////
			assert := assert.New(t)

			agentStorage := NewAgent()
			err := agentStorage.ImportFromJSON([]byte(tt.metrics))
			assert.NoError(err)

			agentStorage.SendMetrics(cfg, logger)

			serverJSON, err := mockStorage.ExportToJSON(context.TODO())
			assert.NoError(err)

			assert.JSONEq(string(agentStorage.ExportToJSON()), string(serverJSON))
		})
	}
}

func TestMetrics_SendMetricsBatch(t *testing.T) {
	cfg := &config.Config{}
	logger := logging.GetLogger()
	cfg.Settings.Address = "http://localhost"

	tests := []struct {
		name    string
		metrics string
	}{
		{
			name:    "case #1 / positive",
			metrics: `[{"id":"test1","type":"counter","delta":12},{"id":"test2","type":"gauge","value":12.1}]`,
		},
		{
			name:    "case #2 / positive",
			metrics: `[{"id":"test1","type":"counter","delta":12},{"id":"test1","type":"counter","delta":12}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			mockStorage := cache.NewMemStorage()

			httpmock.RegisterResponder(http.MethodPost, "http://localhost/updates/",
				func(req *http.Request) (*http.Response, error) {
					body, err := io.ReadAll(req.Body)
					if err != nil {
						return httpmock.NewStringResponse(http.StatusBadRequest, ""), err
					}

					err = mockStorage.ImportFromJSON(req.Context(), body)
					if err != nil {
						return httpmock.NewStringResponse(http.StatusBadRequest, ""), err
					}

					resp := httpmock.NewStringResponse(200, "")
					return resp, nil
				})

			//////////
			assert := assert.New(t)

			agentStorage := NewAgent()
			err := agentStorage.ImportFromJSON([]byte(tt.metrics))
			assert.NoError(err)

			agentStorage.SendMetrics(cfg, logger)

			serverJSON, err := mockStorage.ExportToJSON(context.TODO())
			assert.NoError(err)

			assert.JSONEq(string(agentStorage.ExportToJSON()), string(serverJSON))
		})
	}
}
