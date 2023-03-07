package metric

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	GaugeType   = "gauge"
	CounterType = "counter"
)

var ErrNoResult = errors.New("metric not found")
var ErrWrongHash = errors.New("wrong hash for metric")

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type MetricType interface {
	int64 | float64
}

// NewMetric создает метрику
// В качестве значения можно передать типы float64 и int64
func NewMetric[T MetricType](name, metricType string, value T) Metric {
	var metric Metric
	metric.ID = name
	metric.MType = metricType
	switch metricType {
	case CounterType:
		val := any(value).(int64)
		metric.Delta = &val
	case GaugeType:
		val := any(value).(float64)
		metric.Value = &val
	}
	return metric
}

// Метод Marshal используется для сериализации метрики в формат JSON.
// Функция возвращает срез байтов, содержащий сериализованную метрику.
func (m Metric) Marshal() []byte {
	data, _ := json.Marshal(m)
	return data
}

// GetHash возвращает SHA-256 HMAC хэш строки, которая представляет данные метрики.
// Хэш генерируется с использованием секретного ключа, передаваемого в качестве аргумента функции.
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

// Функция IsValidHash проверяет, совпадает ли HMAC-хэш,
// вычисленный с использованием переданного ключа, с HMAC-хэшем, сохраненным в метрике.
func (m *Metric) IsValidHash(key string) bool {
	return m.GetHash(key) == m.Hash
}
