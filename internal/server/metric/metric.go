package metric

import (
	"encoding/json"
	"log"
	"sync"
)

type Storage interface {
	SetStorage(storage Storage)
	UpdateGaugeElem(name string, value float64)
	UpdateCounterElem(name string, value int64)
	FindCounterByName(name string) (int64, bool)
	FindGaugeByName(name string) (float64, bool)
	FindAll() MemStorage
	ExportToJSON() string
}

const (
	GaugeType   = "gauge"
	CounterType = "counter"
)

type MemStorage struct {
	GaugeMutex   *sync.RWMutex
	GaugeMetrics map[string]float64 `json:"gauge_metrics,omitempty"`

	CounterMutex   *sync.RWMutex
	CounterMetrics map[string]int64 `json:"counter_metrics,omitempty"`
}

func NewMemStorage() Storage {
	return &MemStorage{
		GaugeMutex:     &sync.RWMutex{},
		CounterMutex:   &sync.RWMutex{},
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func (m *MemStorage) SetStorage(storage Storage) {
	m2 := storage.FindAll()

	m.GaugeMutex.Lock()
	m.GaugeMetrics = m2.GaugeMetrics
	m.GaugeMutex.Unlock()

	m.CounterMutex.Lock()
	m.CounterMetrics = m2.CounterMetrics
	m.CounterMutex.Unlock()
}

func (m *MemStorage) UpdateGaugeElem(name string, value float64) {
	m.GaugeMutex.Lock()
	defer m.GaugeMutex.Unlock()
	m.GaugeMetrics[name] = value
}

func (m *MemStorage) UpdateCounterElem(name string, value int64) {
	m.CounterMutex.Lock()
	defer m.CounterMutex.Unlock()
	m.CounterMetrics[name] += value
}

func (m *MemStorage) FindGaugeByName(name string) (float64, bool) {
	m.GaugeMutex.RLock()
	defer m.GaugeMutex.RUnlock()
	v, ok := m.GaugeMetrics[name]
	return v, ok
}

func (m *MemStorage) FindCounterByName(name string) (int64, bool) {
	m.CounterMutex.RLock()
	defer m.CounterMutex.RUnlock()
	v, ok := m.CounterMetrics[name]
	return v, ok
}

func (m *MemStorage) FindAll() MemStorage {
	m.GaugeMutex.RLock()
	defer m.GaugeMutex.RUnlock()
	m.CounterMutex.RLock()
	defer m.CounterMutex.RUnlock()

	return MemStorage{
		GaugeMetrics:   m.GaugeMetrics,
		CounterMetrics: m.CounterMetrics,
	}
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func MetricToJSON(name, metricType string, value interface{}) string {
	m := make(map[string]interface{})
	m["id"] = name
	m["type"] = metricType
	switch metricType {
	case GaugeType:
		m["value"] = value
	case CounterType:
		m["delta"] = value
	default:
		log.Fatal("MetricToJSON wrong type")
	}

	ans, _ := json.Marshal(m)

	return string(ans)
}

func (m *MemStorage) ExportToJSON() string {
	encoded := "["

	m.GaugeMutex.RLock()
	defer m.GaugeMutex.RUnlock()
	m.CounterMutex.RLock()
	defer m.CounterMutex.RUnlock()

	valueCount := len(m.CounterMetrics) + len(m.GaugeMetrics)
	iterCount := 0
	for k, v := range m.GaugeMetrics {
		iterCount++
		encoded += MetricToJSON(k, GaugeType, v)
		if valueCount == iterCount {
			break
		}
		encoded += ",\n"
	}

	for k, v := range m.CounterMetrics {
		iterCount++
		encoded += MetricToJSON(k, CounterType, v)
		if valueCount == iterCount {
			break
		}
		encoded += ",\n"
	}

	encoded += "]"
	return encoded
}
