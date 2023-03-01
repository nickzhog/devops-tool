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
	GaugeMetrics   map[string]float64 `json:"gauge_metrics,omitempty"`
	CounterMetrics map[string]int64   `json:"counter_metrics,omitempty"`
}

func NewMemStorage() *memStorage {
	return &memStorage{
		mutex:          new(sync.RWMutex),
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func (m *memStorage) UpsertMetric(ctx context.Context, metricElem *metric.Metric) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	switch metricElem.MType {
	case metric.GaugeType:
		m.GaugeMetrics[metricElem.ID] = *metricElem.Value
	case metric.CounterType:
		delta := *metricElem.Delta
		oldDelta, exist := m.CounterMetrics[metricElem.ID]
		if exist {
			delta += oldDelta
			metricElem.Delta = &delta
		}
		m.CounterMetrics[metricElem.ID] = delta

	default:
		return errors.New("wrong metric type")
	}

	return nil
}

func (m *memStorage) FindMetric(ctx context.Context, name, mtype string) (metricElem metric.Metric, err error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var (
		ok    bool
		value float64
		delta int64
	)

	switch mtype {
	case metric.GaugeType:
		value, ok = m.GaugeMetrics[name]
		metricElem.Value = &value
	case metric.CounterType:
		delta, ok = m.CounterMetrics[name]
		metricElem.Delta = &delta
	}

	if !ok {
		return metric.Metric{}, metric.ErrNoResult
	}

	return
}

func (m *memStorage) ExportToJSON(ctx context.Context) ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var metrics []metric.Metric
	for k, v := range m.GaugeMetrics {
		metrics = append(metrics, metric.NewMetric(k, metric.GaugeType, v))
	}

	for k, v := range m.CounterMetrics {
		metrics = append(metrics, metric.NewMetric(k, metric.CounterType, v))
	}

	if len(metrics) < 1 {
		return []byte(`[]`), nil
	}

	encoded, err := json.Marshal(metrics)
	if err != nil {
		return []byte(``), err
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
		err = m.UpsertMetric(ctx, &v)
		if err != nil {
			return err
		}
	}

	return nil
}
