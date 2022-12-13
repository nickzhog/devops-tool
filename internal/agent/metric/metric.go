package metric

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"sync"

	"github.com/nickzhog/practicum-metric/internal/agent/config"
	"github.com/nickzhog/practicum-metric/internal/server/metric"
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
	var url string
	var answer []byte
	var err error
	m.GaugeMutex.RLock()
	for k, v := range m.GaugeMetrics {
		url = fmt.Sprintf("%s/update/gauge/%s/%v", cfg.Settings.Address, k, v)

		_, _ = sendRequest(url, "", http.MethodGet)

		////

		url = fmt.Sprintf("%s/update", cfg.Settings.Address)
		body := metric.MetricToJSON(k, metric.GaugeType, v)
		answer, err = sendRequest(url, body, http.MethodPost)
	}
	m.GaugeMutex.RUnlock()

	m.CounterMutex.RLock()
	for k, v := range m.CounterMetrics {
		url = fmt.Sprintf("%s/update/counter/%s/%v", cfg.Settings.Address, k, v)

		_, _ = sendRequest(url, "", http.MethodGet)

		////

		url = fmt.Sprintf("%s/update", cfg.Settings.Address)
		body := metric.MetricToJSON(k, metric.CounterType, v)
		answer, err = sendRequest(url, body, http.MethodPost)
	}
	m.CounterMutex.RUnlock()
	logger.Tracef("metrics sended to: %s, last err: %v, last answer: %v", cfg.Settings.Address, err, string(answer))
}
