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
			metric:     metric.NewCounterMetric("good_counter", 10),
			wantResult: int64(10),
		},
		{
			name:       "increment test",
			metric:     metric.NewCounterMetric("good_counter", 10),
			wantResult: int64(20),
		},
		{
			name:       "gauge metric",
			metric:     metric.NewGaugeMetric("good_gauge", 10),
			wantResult: float64(10),
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			err := storage.UpsertMetric(ctx, tt.metric)
			assert.NoError(err)

			metricElem, err := storage.FindMetric(ctx, tt.metric.ID, tt.metric.MType)

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

func BenchmarkExportToJSON(b *testing.B) {
	metrics := []byte(`
	[
		{"id":"good_gauge","type":"gauge","value":321},
		{"id":"good_gauge2","type":"gauge","value":123},
		{"id":"good_counter","type":"counter","delta":10},
		{"id":"good_counter2","type":"counter","delta":14},
		{"id":"good_gauge3","type":"gauge","value":321},
		{"id":"good_gauge4","type":"gauge","value":321},
		{"id":"good_gauge5","type":"gauge","value":321},
		{"id":"good_gauge6","type":"gauge","value":321},
		{"id":"good_gauge7","type":"gauge","value":321},
		{"id":"good_gauge8","type":"gauge","value":321},
		{"id":"good_gauge9","type":"gauge","value":321},
		{"id":"good_gauge10","type":"gauge","value":321},
		{"id":"good_gauge11","type":"gauge","value":321},
		{"id":"good_gauge12","type":"gauge","value":321}
	]`)

	assert := assert.New(b)

	storage := NewMemStorage()
	err := storage.ImportFromJSON(context.Background(), metrics)
	assert.NoError(err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.ReportAllocs()
		_, err := storage.ExportToJSON(context.Background())
		assert.NoError(err)
	}
}
