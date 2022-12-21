package metric

import (
	"encoding/json"
	"sync"
)

type Storage interface {
	UpdateGaugeElem(name string, value float64)
	UpdateCounterElem(name string, value int64)
	FindCounterByName(name string) (int64, bool)
	FindGaugeByName(name string) (float64, bool)
	FindAll() MemStorage
	ExportToJSON() []byte
}

const (
	GaugeType   = "gauge"
	CounterType = "counter"
)

type MemStorage struct {
	gaugeMutex   *sync.RWMutex
	GaugeMetrics map[string]float64 `json:"gauge_metrics,omitempty"`

	counterMutex   *sync.RWMutex
	CounterMetrics map[string]int64 `json:"counter_metrics,omitempty"`
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gaugeMutex:     &sync.RWMutex{},
		GaugeMetrics:   make(map[string]float64),
		counterMutex:   &sync.RWMutex{},
		CounterMetrics: make(map[string]int64),
	}
}

func (m *MemStorage) UpdateGaugeElem(name string, value float64) {
	m.gaugeMutex.Lock()
	defer m.gaugeMutex.Unlock()
	m.GaugeMetrics[name] = value
}

func (m *MemStorage) UpdateCounterElem(name string, value int64) {
	m.counterMutex.Lock()
	defer m.counterMutex.Unlock()
	m.CounterMetrics[name] += value
}

func (m *MemStorage) FindGaugeByName(name string) (float64, bool) {
	m.gaugeMutex.RLock()
	defer m.gaugeMutex.RUnlock()
	v, ok := m.GaugeMetrics[name]
	return v, ok
}

func (m *MemStorage) FindCounterByName(name string) (int64, bool) {
	m.counterMutex.RLock()
	defer m.counterMutex.RUnlock()
	v, ok := m.CounterMetrics[name]
	return v, ok
}

func (m MemStorage) FindAll() MemStorage {
	m.gaugeMutex.RLock()
	defer m.gaugeMutex.RUnlock()
	m.counterMutex.RLock()
	defer m.counterMutex.RUnlock()

	return MemStorage{
		GaugeMetrics:   m.GaugeMetrics,
		CounterMetrics: m.CounterMetrics,
	}
}

type MetricsExport struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func MetricToJSON(name, metricType string, value interface{}) []byte {
	m := make(map[string]interface{})
	m["id"] = name
	m["type"] = metricType
	switch metricType {
	case GaugeType:
		m["value"] = value
	case CounterType:
		m["delta"] = value
	}

	ans, _ := json.Marshal(m)

	return ans
}

func (m *MemStorage) ExportToJSON() []byte {
	encoded := "["
	m.gaugeMutex.RLock()
	defer m.gaugeMutex.RUnlock()
	m.counterMutex.RLock()
	defer m.counterMutex.RUnlock()

	valueCount := len(m.CounterMetrics) + len(m.GaugeMetrics)
	iterCount := 0
	for k, v := range m.GaugeMetrics {
		iterCount++
		encoded += string(MetricToJSON(k, GaugeType, v))
		if valueCount == iterCount {
			break
		}
		encoded += ",\n"
	}

	for k, v := range m.CounterMetrics {
		iterCount++
		encoded += string(MetricToJSON(k, CounterType, v))
		if valueCount == iterCount {
			break
		}
		encoded += ",\n"
	}

	encoded += "]"

	return []byte(encoded)
}
