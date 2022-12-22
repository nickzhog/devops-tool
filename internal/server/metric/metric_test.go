package metric

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorage_UpdateCounter(t *testing.T) {

	storage := &MemStorage{
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.UpdateCounter(tt.setName, tt.setValue)
			val, ok := storage.FindCounterByName(tt.setName)
			assert := assert.New(t)
			assert.Equal(tt.wantResult, val)
			assert.True(ok)
		})
	}
}
