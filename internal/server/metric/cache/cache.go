package cache

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/nickzhog/devops-tool/internal/server/metric"
)

type memStorage struct {
	mutex          *sync.RWMutex
	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
}

func NewMemStorage() *memStorage {
	return &memStorage{
		mutex:          new(sync.RWMutex),
		gaugeMetrics:   make(map[string]float64),
		counterMetrics: make(map[string]int64),
	}
}

func (m *memStorage) Ping(ctx context.Context) error {
	return nil
}

func (m *memStorage) UpsertMetric(ctx context.Context, metricElem metric.Metric) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	switch metricElem.MType {
	case metric.GaugeType:
		m.gaugeMetrics[metricElem.ID] = *metricElem.Value
	case metric.CounterType:
		delta := *metricElem.Delta
		oldDelta, exist := m.counterMetrics[metricElem.ID]
		if exist {
			delta += oldDelta
			metricElem.Delta = &delta
		}
		m.counterMetrics[metricElem.ID] = delta

	default:
		return errors.New("wrong metric type")
	}

	return nil
}

func (m *memStorage) FindMetric(ctx context.Context, name, mtype string) (metric.Metric, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var (
		answer = metric.Metric{ID: name, MType: mtype}
		ok     bool
		value  float64
		delta  int64
	)

	switch mtype {
	case metric.GaugeType:
		value, ok = m.gaugeMetrics[name]
		answer.Value = &value
	case metric.CounterType:
		delta, ok = m.counterMetrics[name]
		answer.Delta = &delta
	}

	if !ok {
		return metric.Metric{}, metric.ErrNoResult
	}

	return answer, nil
}

func (m *memStorage) ExportToJSON(ctx context.Context) ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var metrics []metric.Metric
	for k, v := range m.gaugeMetrics {
		metrics = append(metrics, metric.NewGaugeMetric(k, v))
	}

	for k, v := range m.counterMetrics {
		metrics = append(metrics, metric.NewCounterMetric(k, v))
	}

	if len(metrics) < 1 {
		return []byte(`[]`), nil
	}

	encoded, err := json.Marshal(metrics)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func (m *memStorage) ImportFromJSON(ctx context.Context, data []byte) error {
	var metrics []metric.Metric
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return err
	}

	for _, v := range metrics {
		err = m.UpsertMetric(ctx, v)
		if err != nil {
			return err
		}
	}

	return nil
}
