package cache

import (
	"context"
	"sync"
	"testing"

	"github.com/nickzhog/devops-tool/internal/server/metric"
	"github.com/stretchr/testify/assert"
)

func TestMemStorage_Upsert(t *testing.T) {
	storage := &memStorage{
		mutex:          &sync.RWMutex{},
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
	storage.CounterMetrics["good_counter"] = 10

	tests := []struct {
		name       string
		setName    string
		setValue   int64
		wantResult int64
	}{
		{
			name:       "increment test",
			setName:    "good_counter",
			setValue:   10,
			wantResult: 20,
		},
		{
			name:       "new value",
			setName:    "good_counter2",
			setValue:   10,
			wantResult: 10,
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metricElem := metric.NewMetric(tt.setName, metric.CounterType, tt.setValue)
			storage.UpsertMetric(ctx, &metricElem)
			metricElem, ok := storage.FindMetric(ctx, tt.setName, metric.CounterType)
			assert := assert.New(t)
			assert.Equal(tt.wantResult, *metricElem.Delta)
			assert.True(ok)
		})
	}
}
