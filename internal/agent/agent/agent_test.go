package agent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/nickzhog/devops-tool/internal/agent/config"
	"github.com/nickzhog/devops-tool/internal/server/service/cache"
	"github.com/nickzhog/devops-tool/pkg/logging"
	"github.com/nickzhog/devops-tool/pkg/metric"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_SendMetrics(t *testing.T) {
	cfg := &config.Config{}
	logger := logging.GetLogger()
	cfg.Settings.Address = "http://localhost"

	tests := []struct {
		name     string
		metrics  string
		whantErr bool
	}{
		{
			name:     "case #1 / positive",
			metrics:  `[{"id":"test1","type":"counter","delta":12},{"id":"test2","type":"gauge","value":12.1}]`,
			whantErr: false,
		},
		{
			name:     "case #2 / positive",
			metrics:  `[{"id":"test1","type":"counter","delta":12},{"id":"test1","type":"counter","delta":12}]`,
			whantErr: false,
		},
		{
			name:     "case #3 / bad json",
			metrics:  `[{"id":"test1","type"]`,
			whantErr: true,
		},
		{
			name:     "case #4 / wrong metric",
			metrics:  `[{"id":"test1","type":"wrong_type","delta":12},{"id":"test1","type":"wrong_type","delta":12}]`,
			whantErr: true,
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

					var elem metric.Metric
					err = json.Unmarshal(body, &elem)
					if err != nil {
						return httpmock.NewStringResponse(http.StatusBadRequest, ""), err
					}

					err = mockStorage.UpsertMetric(req.Context(), elem)
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

			agentStorage := NewAgent(cfg, logger)
			err := agentStorage.ImportFromJSON([]byte(tt.metrics))
			if tt.whantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)

			agentStorage.SendMetricsHTTP(context.Background())

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
		name     string
		metrics  string
		whantErr bool
	}{
		{
			name:    "case #1 / positive",
			metrics: `[{"id":"test1","type":"counter","delta":12},{"id":"test2","type":"gauge","value":12.1}]`,
		},
		{
			name:    "case #2 / positive",
			metrics: `[{"id":"test1","type":"counter","delta":12},{"id":"test1","type":"counter","delta":12}]`,
		},
		{
			name:     "case #3 / bad json",
			metrics:  `[{"id":"test1","type"]`,
			whantErr: true,
		},
		{
			name:     "case #4 / wrong metric",
			metrics:  `[{"id":"test1","type":"wrong_type","delta":12},{"id":"test1","type":"wrong_type","delta":12}]`,
			whantErr: true,
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

			agentStorage := NewAgent(cfg, logger)
			err := agentStorage.ImportFromJSON([]byte(tt.metrics))
			if tt.whantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)

			agentStorage.SendMetricsHTTP(context.Background())

			serverJSON, err := mockStorage.ExportToJSON(context.TODO())
			assert.NoError(err)

			assert.JSONEq(string(agentStorage.ExportToJSON()), string(serverJSON))
		})
	}
}
