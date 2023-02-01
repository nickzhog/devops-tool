package metric

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type Storage interface {
	UpsertMetric(ctx context.Context, metrics *Metric) error
	FindMetric(ctx context.Context, name, mtype string) (Metric, bool)
	ExportToJSON(ctx context.Context) ([]byte, error)
	ImportFromJSON(ctx context.Context, data []byte) error
}

const (
	GaugeType   = "gauge"
	CounterType = "counter"
)

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

func NewMetric(name, metricType string, value interface{}) Metric {
	var metric Metric
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

func (m Metric) Marshal() []byte {
	data, _ := json.Marshal(m)
	return data
}

func (m *Metric) GetHash(key string) string {
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

func (m *Metric) IsValidHash(key string) bool {
	return m.GetHash(key) == m.Hash
}
