package metric

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
)

type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	FindCounterByName(name string) (int64, bool)
	FindGaugeByName(name string) (float64, bool)
	ExportToJSON() []byte
}

const (
	GaugeType   = "gauge"
	CounterType = "counter"
)

type MemStorage struct {
	mutex          *sync.RWMutex
	GaugeMetrics   map[string]float64 `json:"gauge_metrics,omitempty"`
	CounterMetrics map[string]int64   `json:"counter_metrics,omitempty"`
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		mutex:          &sync.RWMutex{},
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.GaugeMetrics[name] = value
}

func (m *MemStorage) UpdateCounter(name string, value int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.CounterMetrics[name] += value
}

func (m *MemStorage) FindGaugeByName(name string) (float64, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	v, ok := m.GaugeMetrics[name]
	return v, ok
}

func (m *MemStorage) FindCounterByName(name string) (int64, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	v, ok := m.CounterMetrics[name]
	return v, ok
}

type MetricExport struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

func (m MetricExport) Marshal() []byte {
	data, _ := json.Marshal(m)
	return data
}

func (m *MetricExport) GetHash(key string) string {
	var data string
	switch m.MType {
	case GaugeType:
		data = fmt.Sprintf("%s:%s:%f", m.ID, GaugeType, *m.Value)
	case CounterType:
		data = fmt.Sprintf("%s:%s:%d", m.ID, CounterType, *m.Delta)
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func MetricToExport(name, metricType string, value interface{}) MetricExport {
	var metric MetricExport
	metric.ID = name
	metric.MType = metricType
	switch metricType {
	case CounterType:
		val := value.(int64)
		metric.Delta = &val
	case GaugeType:
		val := value.(float64)
		metric.Value = &val
	}
	return metric
}

func (m *MemStorage) ExportToJSON() []byte {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var metrics []MetricExport
	for k, v := range m.GaugeMetrics {
		metrics = append(metrics, MetricToExport(k, GaugeType, v))
	}

	for k, v := range m.CounterMetrics {
		metrics = append(metrics, MetricToExport(k, CounterType, v))
	}

	encoded, _ := json.Marshal(metrics)

	return encoded
}
