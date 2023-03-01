package cache

import (
	"context"
	"testing"

	"github.com/nickzhog/devops-tool/internal/server/metric"
	"github.com/stretchr/testify/assert"
)

func TestMemStorage_Upsert(t *testing.T) {
	storage := NewMemStorage()

	tests := []struct {
		name       string
		metric     metric.Metric
		wantResult interface{}
	}{
		{
			name:       "counter metric",
			metric:     metric.NewMetric("good_counter", metric.CounterType, int64(10)),
			wantResult: int64(10),
		},
		{
			name:       "increment test",
			metric:     metric.NewMetric("good_counter", metric.CounterType, int64(10)),
			wantResult: int64(20),
		},
		{
			name:       "gauge metric",
			metric:     metric.NewMetric("good_gauge", metric.GaugeType, float64(10)),
			wantResult: float64(10),
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			storage.UpsertMetric(ctx, &tt.metric)

			metricElem, err := storage.FindMetric(ctx, tt.metric.ID, tt.metric.MType)
			assert := assert.New(t)

			switch tt.metric.MType {
			case metric.CounterType:
				assert.Equal(tt.wantResult, *metricElem.Delta)
			case metric.GaugeType:
				assert.Equal(tt.wantResult, *metricElem.Value)
			}

			assert.NoError(err)
		})
	}
}
