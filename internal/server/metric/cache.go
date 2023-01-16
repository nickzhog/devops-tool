package metric

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
)

type memStorage struct {
	mutex          *sync.RWMutex
	GaugeMetrics   map[string]float64 `json:"gauge_metrics,omitempty"`
	CounterMetrics map[string]int64   `json:"counter_metrics,omitempty"`
}

func NewMemStorage() Storage {
	return &memStorage{
		mutex:          new(sync.RWMutex),
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func (m *memStorage) UpsertMetric(ctx context.Context, metric *Metric) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	switch metric.MType {
	case GaugeType:
		m.GaugeMetrics[metric.ID] = *metric.Value
	case CounterType:
		delta := *metric.Delta
		oldDelta, exist := m.CounterMetrics[metric.ID]
		if exist {
			delta += oldDelta
			metric.Delta = &delta
		}
		m.CounterMetrics[metric.ID] = delta

	default:
		return errors.New("wrong metric type")
	}

	return nil
}

func (m *memStorage) FindMetric(ctx context.Context, name, mtype string) (Metric, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var (
		val interface{}
		ok  bool
	)
	switch mtype {
	case GaugeType:
		val, ok = m.GaugeMetrics[name]
	case CounterType:
		val, ok = m.CounterMetrics[name]
	default:
		return Metric{}, false
	}
	if !ok {
		return Metric{}, false
	}

	return NewMetric(name, mtype, val), true
}

func (m *memStorage) ExportToJSON(ctx context.Context) ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var metrics []Metric
	for k, v := range m.GaugeMetrics {
		metrics = append(metrics, NewMetric(k, GaugeType, v))
	}

	for k, v := range m.CounterMetrics {
		metrics = append(metrics, NewMetric(k, CounterType, v))
	}

	encoded, err := json.Marshal(metrics)

	return encoded, err
}

func (m *memStorage) ImportFromJSON(ctx context.Context, data []byte) error {
	var metrics []Metric
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return err
	}

	for _, v := range metrics {
		err = m.UpsertMetric(ctx, &v)
		if err != nil {
			return err
		}
	}

	return nil
}
