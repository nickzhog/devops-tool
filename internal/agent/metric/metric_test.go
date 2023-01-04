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
	"github.com/nickzhog/devops-tool/pkg/logging"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_SendMetrics(t *testing.T) {
	cfg := &config.Config{}
	logger := logging.GetLogger()
	cfg.Settings.Address = "http://localhost"

	tests := []struct {
		name    string
		metrics Agent
	}{
		{
			name: "case #1",
			metrics: Agent{
				CounterMetrics: map[string]int64{
					"good_counter": 10,
				},
			},
		},
		{
			name: "case #2",
			metrics: Agent{
				GaugeMetrics: map[string]float64{
					"good_gauge": 15.51,
				},
			},
		},
		{
			name: "case #3",
			metrics: Agent{
				GaugeMetrics: map[string]float64{
					"good_gauge":  25.51,
					"good_gauge2": 2313.51,
				},
				CounterMetrics: map[string]int64{
					"good_counter":  123,
					"good_counter2": 321,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			mockStorage := serverMetric.NewMemStorage()

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
			agentStorage := NewAgent()
			for k, v := range tt.metrics.CounterMetrics {
				agentStorage.CounterMetrics[k] = v
			}
			for k, v := range tt.metrics.GaugeMetrics {
				agentStorage.GaugeMetrics[k] = v
			}

			agentStorage.SendMetrics(cfg, logger)

			assert := assert.New(t)
			assert.Equal(len(agentStorage.CounterMetrics)+len(agentStorage.GaugeMetrics),
				httpmock.GetTotalCallCount())

			serverJSON, _ := mockStorage.ExportToJSON(context.TODO())
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
		metrics Agent
	}{
		{
			name: "case #1",
			metrics: Agent{
				CounterMetrics: map[string]int64{
					"good_counter": 10,
				},
			},
		},
		{
			name: "case #2",
			metrics: Agent{
				GaugeMetrics: map[string]float64{
					"good_gauge": 15.51,
				},
			},
		},
		{
			name: "case #3",
			metrics: Agent{
				GaugeMetrics: map[string]float64{
					"good_gauge":  25.51,
					"good_gauge2": 321.123,
				},
				CounterMetrics: map[string]int64{
					"good_counter": 321,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			mockStorage := serverMetric.NewMemStorage()

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
			agentStorage := NewAgent()
			for k, v := range tt.metrics.CounterMetrics {
				agentStorage.CounterMetrics[k] = v
			}
			for k, v := range tt.metrics.GaugeMetrics {
				agentStorage.GaugeMetrics[k] = v
			}

			agentStorage.SendMetrics(cfg, logger)

			assert := assert.New(t)

			serverJSON, _ := mockStorage.ExportToJSON(context.TODO())
			assert.JSONEq(string(agentStorage.ExportToJSON()), string(serverJSON))
		})
	}
}
