package metric

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"sync"

	"github.com/nickzhog/devops-tool/internal/agent/config"
	"github.com/nickzhog/devops-tool/internal/server/metric"
	"github.com/nickzhog/devops-tool/pkg/logging"
	"github.com/shirou/gopsutil/mem"
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

	//gopsutil
	mem, _ := mem.VirtualMemory()

	a.GaugeMetrics["CPUutilization1"] = float64(mem.UsedPercent)
	a.GaugeMetrics["TotalMemory"] = float64(mem.Total)
	a.GaugeMetrics["FreeMemory"] = float64(mem.Free)

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

		sendRequest(url, []byte(``), http.MethodPost)

		////

		url = fmt.Sprintf("%s/update", cfg.Settings.Address)
		metric := metric.NewMetric(k, metric.GaugeType, v)
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
		metric := metric.NewMetric(k, metric.CounterType, v)
		if cfg.Settings.Key != "" {
			metric.Hash = string(metric.GetHash(cfg.Settings.Key))
		}
		body, _ := json.Marshal(metric)
		answer, err = sendRequest(url, body, http.MethodPost)
	}
	logger.Tracef("metrics sended to: %s, last err: %v, last answer: %v", cfg.Settings.Address, err, string(answer))

	if len(jsonData) > 0 {
		url := fmt.Sprintf("%s/updates/", cfg.Settings.Address)
		_, err = sendRequest(url, jsonData, http.MethodPost)
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
		}
	}

	return nil
}

func (a *agent) ExportToJSON() []byte {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	metrics := make([]metric.Metric, 0)
	for k, v := range a.GaugeMetrics {
		metrics = append(metrics, metric.NewMetric(k, metric.GaugeType, v))
	}
	for k, v := range a.CounterMetrics {
		metrics = append(metrics, metric.NewMetric(k, metric.CounterType, v))
	}
	ans, err := json.Marshal(metrics)
	if err != nil {
		return []byte(``)
	}
	return ans
}
