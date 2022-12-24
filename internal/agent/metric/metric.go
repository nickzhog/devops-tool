package metric

import (
	"encoding/json"
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

type Agent struct {
	mutex          *sync.RWMutex
	GaugeMetrics   map[string]float64 `json:"gauge_metrics,omitempty"`
	CounterMetrics map[string]int64   `json:"counter_metrics,omitempty"`
}

func NewAgent() *Agent {
	return &Agent{
		mutex:          &sync.RWMutex{},
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func (a *Agent) UpdateMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	v := reflect.ValueOf(memStats)
	typeOfS := v.Type()

	a.mutex.Lock()
	defer a.mutex.Unlock()

	for i := 0; i < v.NumField(); i++ {
		i := i
		floatValue, ok := getFloat(v.Field(i).Interface())
		if !ok {
			continue
		}

		// fmt.Printf("%v float64\n", typeOfS.Field(i).Name)
		a.GaugeMetrics[typeOfS.Field(i).Name] = floatValue
	}
	a.GaugeMetrics["RandomValue"] = float64(rand.Int63n(1000))

	a.CounterMetrics["PollCount"]++
}

func (a *Agent) SendMetrics(cfg *config.Config, logger *logging.Logger) {
	var url string
	var answer []byte
	var err error
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	for k, v := range a.GaugeMetrics {
		url = fmt.Sprintf("%s/update/gauge/%s/%v", cfg.Settings.Address, k, v)

		sendRequest(url, []byte(``), http.MethodPost)

		////

		url = fmt.Sprintf("%s/update", cfg.Settings.Address)
		metric := metric.MetricToExport(k, metric.GaugeType, v)
		if cfg.Settings.Key != "" {
			metric.Hash = string(metric.GetHash(cfg.Settings.Key))
		}
		body, _ := json.Marshal(metric)
		answer, err = sendRequest(url, body, http.MethodPost)
	}

	for k, v := range a.CounterMetrics {
		url = fmt.Sprintf("%s/update/counter/%s/%v", cfg.Settings.Address, k, v)

		sendRequest(url, []byte(``), http.MethodPost)

		////

		url = fmt.Sprintf("%s/update", cfg.Settings.Address)
		metric := metric.MetricToExport(k, metric.CounterType, v)
		if cfg.Settings.Key != "" {
			metric.Hash = string(metric.GetHash(cfg.Settings.Key))
		}
		body, _ := json.Marshal(metric)
		answer, err = sendRequest(url, body, http.MethodPost)
	}

	logger.Tracef("metrics sended to: %s, last err: %v, last answer: %v", cfg.Settings.Address, err, string(answer))
}
