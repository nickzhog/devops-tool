package metric

import (
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
	storage := NewAgent()
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

					var elem serverMetric.MetricExport
					err = json.Unmarshal(body, &elem)
					if err != nil {
						return httpmock.NewStringResponse(http.StatusBadRequest, ""), err
					}
					var value interface{}

					switch elem.MType {
					case serverMetric.CounterType:
						mockStorage.UpdateCounter(elem.ID, *elem.Delta)
						value = *elem.Delta
					case serverMetric.GaugeType:
						mockStorage.UpdateGauge(elem.ID, *elem.Value)
						value = *elem.Value
					default:
						return httpmock.NewStringResponse(http.StatusBadRequest, ""), err
					}

					output := serverMetric.MetricToExport(elem.ID, elem.MType, value).Marshal()
					resp, err := httpmock.NewJsonResponse(200, output)
					if err != nil {
						return httpmock.NewStringResponse(http.StatusBadGateway, ""), err
					}
					return resp, nil
				})

			//////////
			assert := assert.New(t)

			storage.CounterMetrics = tt.metrics.CounterMetrics
			storage.GaugeMetrics = tt.metrics.GaugeMetrics

			storage.SendMetrics(cfg, logger)

			assert.Equal(len(storage.CounterMetrics)+len(storage.GaugeMetrics),
				httpmock.GetTotalCallCount())
			if len(storage.CounterMetrics) > 0 {
				assert.Equal(storage.CounterMetrics, mockStorage.CounterMetrics)
			}
			if len(storage.GaugeMetrics) > 0 {
				assert.Equal(storage.GaugeMetrics, mockStorage.GaugeMetrics)
			}
		})
	}
}
