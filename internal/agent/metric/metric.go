package metric

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"sync"

	"github.com/nickzhog/practicum-metric/internal/agent/config"
	"github.com/nickzhog/practicum-metric/pkg/logging"
)

type Metrics struct {
	GaugeMutex   *sync.RWMutex
	GaugeMetrics map[string]float64 `json:"gauge_metrics,omitempty"`

	CounterMutex   *sync.RWMutex
	CounterMetrics map[string]int64 `json:"counter_metrics,omitempty"`
}

func (m *Metrics) InitMetrics() {
	m.GaugeMutex = &sync.RWMutex{}
	m.GaugeMetrics = make(map[string]float64)
	m.CounterMutex = &sync.RWMutex{}
	m.CounterMetrics = make(map[string]int64)
}

func (m *Metrics) UpdateMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	v := reflect.ValueOf(memStats)
	typeOfS := v.Type()

	for i := 0; i < v.NumField(); i++ {
		i := i
		floatValue, ok := getFloat(v.Field(i).Interface())
		if !ok {
			continue
		}

		// fmt.Printf("%v float64\n", typeOfS.Field(i).Name)
		m.GaugeMutex.Lock()
		m.GaugeMetrics[typeOfS.Field(i).Name] = floatValue
		m.GaugeMutex.Unlock()
	}
	m.GaugeMutex.Lock()
	m.GaugeMetrics["RandomValue"] = float64(rand.Int63n(1000))
	m.GaugeMutex.Unlock()

	m.CounterMutex.Lock()
	m.CounterMetrics["PollCount"]++
	m.CounterMutex.Unlock()
}

func (m *Metrics) SendMetrics(cfg *config.Config, logger *logging.Logger) {
	m.GaugeMutex.RLock()
	for k, v := range m.GaugeMetrics {
		url := fmt.Sprintf("%s/update/gauge/%s/%f", cfg.SendTo.Address, k, v)

		answer, err := sendRequest(url, fmt.Sprintf("%f", v))
		if err != nil {
			logger.Errorf("req err: %s, request: %s", err.Error(), url)
		}
		logger.Tracef("gauge update answer: %s", string(answer))
	}
	m.GaugeMutex.RUnlock()

	m.CounterMutex.RLock()
	for k, v := range m.CounterMetrics {
		url := fmt.Sprintf("%s/update/counter/%s/%v", cfg.SendTo.Address, k, v)

		answer, err := sendRequest(url, fmt.Sprintf("%v", v))
		if err != nil {
			logger.Errorf("req err: %s, request: %s", err.Error(), url)
		}
		logger.Tracef("counter update answer: %s", string(answer))
	}
	m.CounterMutex.RUnlock()
}
