package metric

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/nickzhog/practicum-metric/internal/agent/config"
	serverMetric "github.com/nickzhog/practicum-metric/internal/server/metric"
	"github.com/nickzhog/practicum-metric/pkg/logging"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_SendMetrics(t *testing.T) {
	var storage Metrics
	storage.InitMetrics()
	cfg := &config.Config{}
	logger := logging.GetLogger()
	cfg.Settings.Address = "http://localhost"

	tests := []struct {
		name    string
		metrics Metrics
	}{
		{
			name: "case #1",
			metrics: Metrics{
				CounterMetrics: map[string]int64{
					"good_counter": 10,
				},
			},
		},
		{
			name: "case #2",
			metrics: Metrics{
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

					var elem serverMetric.MetricsExport
					err = json.Unmarshal(body, &elem)
					if err != nil {
						return httpmock.NewStringResponse(http.StatusBadRequest, ""), err
					}
					var value interface{}
					switch elem.MType {
					case serverMetric.CounterType:
						mockStorage.UpdateCounterElem(elem.ID, *elem.Delta)
						value = *elem.Delta
					case serverMetric.GaugeType:
						mockStorage.UpdateGaugeElem(elem.ID, *elem.Value)
						value = *elem.Value
					default:
						return httpmock.NewStringResponse(http.StatusBadRequest, ""), err
					}

					resp, err := httpmock.NewJsonResponse(200,
						serverMetric.MetricToJSON(elem.ID, elem.MType, value))
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
				assert.Equal(storage.CounterMetrics, mockStorage.FindAll().CounterMetrics)
			}
			if len(storage.GaugeMetrics) > 0 {
				assert.Equal(storage.GaugeMetrics, mockStorage.FindAll().GaugeMetrics)
			}
		})
	}
}
