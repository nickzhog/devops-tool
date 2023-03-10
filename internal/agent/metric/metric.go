package metric

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/nickzhog/devops-tool/internal/agent/config"
	"github.com/nickzhog/devops-tool/internal/server/metric"
	"github.com/nickzhog/devops-tool/pkg/logging"
)

type Agent interface {
	UpdateMetrics()
	SendMetrics(cfg *config.Config, logger *logging.Logger)
	ExportToJSON() []byte
	ImportFromJSON(data []byte) error
}

type agent struct {
	mutex          *sync.RWMutex
	GaugeMetrics   map[string]float64 `json:"gauge_metrics,omitempty"`
	CounterMetrics map[string]int64   `json:"counter_metrics,omitempty"`
}

func NewAgent() Agent {
	return &agent{
		mutex:          new(sync.RWMutex),
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func (a *agent) UpdateMetrics() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	setGaugeMetrics(a.GaugeMetrics)

	a.CounterMetrics["PollCount"]++
}

func (a *agent) SendMetrics(cfg *config.Config, logger *logging.Logger) {
	var url string
	var answer []byte
	var err error

	jsonData := a.ExportToJSON()

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	for k, v := range a.GaugeMetrics {
		url = fmt.Sprintf("%s/update/gauge/%s/%v", cfg.Settings.Address, k, v)

		sendRequest(url, nil)

		////

		url = fmt.Sprintf("%s/update", cfg.Settings.Address)
		metric := metric.NewGaugeMetric(k, v)
		if cfg.Settings.Key != "" {
			metric.Hash = string(metric.GetHash(cfg.Settings.Key))
		}
		body, _ := json.Marshal(metric)
		answer, err = sendRequest(url, body)
	}

	for k, v := range a.CounterMetrics {
		url = fmt.Sprintf("%s/update/counter/%s/%v", cfg.Settings.Address, k, v)

		sendRequest(url, nil)

		////

		url = fmt.Sprintf("%s/update", cfg.Settings.Address)
		metric := metric.NewCounterMetric(k, v)
		if cfg.Settings.Key != "" {
			metric.Hash = string(metric.GetHash(cfg.Settings.Key))
		}
		body, _ := json.Marshal(metric)
		answer, err = sendRequest(url, body)
	}

	logger.Tracef("metrics sended to: %s, last err: %v, last answer: %s", cfg.Settings.Address, err, answer)

	if len(jsonData) > 0 {
		url := fmt.Sprintf("%s/updates/", cfg.Settings.Address)
		_, err = sendRequest(url, jsonData)
		if err != nil {
			logger.Error(err)
		}
	}
}

func (a *agent) ImportFromJSON(data []byte) error {
	var metrics []metric.Metric
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return err
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	for _, v := range metrics {
		switch v.MType {
		case metric.CounterType:
			a.CounterMetrics[v.ID] = *v.Delta
		case metric.GaugeType:
			a.GaugeMetrics[v.ID] = *v.Value
		default:
			return fmt.Errorf("wring metric type: %s", v.MType)
		}
	}

	return nil
}

func (a *agent) ExportToJSON() []byte {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	metrics := make([]metric.Metric, 0)
	for k, v := range a.GaugeMetrics {
		metrics = append(metrics, metric.NewGaugeMetric(k, v))
	}
	for k, v := range a.CounterMetrics {
		metrics = append(metrics, metric.NewCounterMetric(k, v))
	}
	ans, err := json.Marshal(metrics)
	if err != nil {
		return nil
	}
	return ans
}
