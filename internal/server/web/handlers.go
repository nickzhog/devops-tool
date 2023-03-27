package web

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/service"
	"github.com/nickzhog/devops-tool/pkg/logging"
	"github.com/nickzhog/devops-tool/pkg/metric"
)

func (h *handler) showError(w http.ResponseWriter, err string, status int) {
	// h.Logger.Error(err)
	m := map[string]string{
		"error": err,
	}
	data, _ := json.Marshal(m)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

type handler struct {
	Storage service.Storage
	Logger  *logging.Logger
	Cfg     *config.Config
}

func NewHandlerData(logger *logging.Logger, cfg *config.Config, storage service.Storage) *handler {
	return &handler{
		Logger:  logger,
		Storage: storage,
		Cfg:     cfg,
	}
}

type ForTemplate struct {
	Key   string
	Value string
}

var templ = template.Must(template.New("index").Parse(
	`
<body>
    <br><br>
    <h1>Gauge:</h1>
    <br><br>
    {{range .GaugeValues}}
        <p> {{.Key}} - {{.Value}}</p>
    <br><br>
    {{end}}
    <br><hr><br>
    <h1>Counter:</h1>
    <br><br>
    {{range .CounterValues}}
        <p> {{.Key}} - {{.Value}}</p>
    <br><br>
    {{end}}
    <br>
</body>	
	`,
))

// Обработчик проверяет доступность хранилища
func (h *handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()
	err := h.Storage.Ping(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(nil)
}

// IndexHandler - главная страница, отображает все доступные метрики и их значения
func (h *handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	data, err := h.Storage.ExportToJSON(r.Context())
	if err != nil {
		h.showError(w, err.Error(), http.StatusInternalServerError)
	}

	var metrics []metric.Metric
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		h.showError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gaugeData := make([]ForTemplate, 0)
	counterData := make([]ForTemplate, 0)
	for _, v := range metrics {
		switch v.MType {
		case metric.CounterType:
			counterData = append(counterData, ForTemplate{Key: v.ID, Value: fmt.Sprintf("%v", *v.Delta)})
		case metric.GaugeType:
			gaugeData = append(gaugeData, ForTemplate{Key: v.ID, Value: fmt.Sprintf("%f", *v.Value)})
		}
	}
	sort.Slice(gaugeData, func(i, j int) bool {
		return gaugeData[i].Key < gaugeData[j].Key
	})
	sort.Slice(counterData, func(i, j int) bool {
		return counterData[i].Key < counterData[j].Key
	})

	m := make(map[string]interface{})

	m["GaugeValues"] = gaugeData
	m["CounterValues"] = counterData

	w.Header().Set("Content-Type", "text/html")
	templ.Execute(w, m)
}

// Обработчик SelectFromBody используется для поиска метрики в хранилище
// на основе данных, переданных в формате JSON в теле HTTP-запроса.
//
// Пример тела запроса:
//
//	{
//		"id": "good_metric",
//		"type": "gauge"
//	}
func (h *handler) SelectFromBody(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.showError(w, "cant get body", http.StatusBadRequest)
		return
	}
	var metricElem metric.Metric
	err = json.Unmarshal(body, &metricElem)
	if err != nil {
		h.showError(w, fmt.Sprintf("cant parse body:%s", string(body)), http.StatusBadRequest)
		return
	}

	metricElem, err = h.Storage.FindMetric(r.Context(), metricElem.ID, metricElem.MType)
	if err != nil {
		if err == metric.ErrNoResult {
			h.showError(w, "not found", http.StatusNotFound)
			return
		}
		h.showError(w, "not found", http.StatusInternalServerError)
	}

	if h.Cfg.Settings.Key != "" {
		metricElem.Hash = string(metricElem.GetHash(h.Cfg.Settings.Key))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(metricElem.Marshal())
}

// Обработчик UpdateFromBody используется для обновления/создания метрики в хранилище
// на основе данных, переданных в формате JSON в теле HTTP-запроса.
//
// Пример тела запроса:
//
//	{
//		"id": "good_metric",
//		"type": "gauge",
//		"value": 10.5
//	}
func (h *handler) UpdateFromBody(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.showError(w, "cant get body", http.StatusBadRequest)
		return
	}
	var metricElem metric.Metric
	err = json.Unmarshal(body, &metricElem)
	if err != nil {
		h.showError(w, fmt.Sprintf("cant parse body:%s", string(body)), http.StatusBadRequest)
		return
	}

	if h.Cfg.Settings.Key != "" {
		if !hmac.Equal(
			[]byte(metricElem.GetHash(h.Cfg.Settings.Key)),
			[]byte(metricElem.Hash)) {
			h.showError(w, "wrong hash", http.StatusBadRequest)
			return
		}
	}

	h.Logger.Tracef("UpdateFromBody: %s", body)

	err = h.Storage.UpsertMetric(r.Context(), metricElem)
	if err != nil {
		h.showError(w, err.Error(), http.StatusBadRequest)
		return
	}

	mcurrent, err := h.Storage.FindMetric(r.Context(), metricElem.ID, metricElem.MType)
	if err != nil {
		h.showError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.Cfg.Settings.Key != "" {
		mcurrent.Hash = string(mcurrent.GetHash(h.Cfg.Settings.Key))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(mcurrent.Marshal())
}

// Обработчик SelectFromURL используется для поиска метрики в хранилище
// на основе данных, переданных в URL-параметрах.
//
// Пример URL-запроса:
// /value/gauge/good_metric
func (h *handler) SelectFromURL(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metric_type")
	if metricType != metric.CounterType && metricType != metric.GaugeType {
		h.showError(w,
			"metric_type is missing in parameters",
			http.StatusBadRequest)
		return
	}

	metricName := chi.URLParam(r, "name")
	if len(metricName) < 1 {
		h.showError(w,
			"name is missing in parameters",
			http.StatusBadRequest)
		return
	}

	metricElem, err := h.Storage.FindMetric(r.Context(), metricName, metricType)
	if err != nil {
		if err == metric.ErrNoResult {
			h.showError(w, "not found", http.StatusNotFound)
			return
		}
		h.showError(w, "not found", http.StatusInternalServerError)
	}

	var v interface{}
	switch metricElem.MType {
	case metric.CounterType:
		v = *metricElem.Delta
	case metric.GaugeType:
		v = *metricElem.Value
	}

	w.Write([]byte(fmt.Sprintf("%v", v)))
}

// Обработчик UpdateFromURL используется для обновления/создания метрики в хранилище
// на основе данных, переданных в URL-параметрах.
//
// Пример URL-запроса:
// /value/gauge/good_metric/10.5
func (h *handler) UpdateFromURL(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metric_type")
	metricName := chi.URLParam(r, "name")
	if len(metricName) < 1 {
		h.showError(w,
			"name is missing in parameters",
			http.StatusBadRequest)
		return
	}

	metricValue := chi.URLParam(r, "value")

	var (
		metricElem  metric.Metric
		valueString string
	)

	switch metricType {
	case metric.GaugeType:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			h.showError(w,
				fmt.Sprintf("cant convert value to float:%v", err.Error()),
				http.StatusBadRequest)
			return
		}
		metricElem = metric.NewGaugeMetric(metricName, value)
		valueString = fmt.Sprintf("%g", value)
	case metric.CounterType:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			h.showError(w,
				fmt.Sprintf("cant convert value to int64:%v", err.Error()),
				http.StatusBadRequest)
			return
		}
		metricElem = metric.NewCounterMetric(metricName, value)
		valueString = fmt.Sprintf("%v", value)

		actualMetric, err := h.Storage.FindMetric(r.Context(), metricName, metricType)
		if err != nil && !errors.Is(err, metric.ErrNoResult) {
			h.showError(w, err.Error(), http.StatusInternalServerError)
			return
		} else if err == nil {
			valueString = fmt.Sprintf("%v", *actualMetric.Delta+value)
		}
	default:
		h.showError(w, "wrong metric type", http.StatusNotImplemented)
		return
	}

	err := h.Storage.UpsertMetric(r.Context(), metricElem)
	if err != nil {
		h.showError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(valueString))
}

// Обработчик UpdateMany используется для обновления/создания множества метрик в хранилище
// на основе данных, переданных в формате JSON в теле HTTP-запроса.
//
// Пример тела запроса:
//
//	[
//		{
//			"id": "good_metric",
//			"type": "gauge",
//			"value": 10.5
//		},
//		{
//			"id": "good_metric2",
//			"type": "counter",
//			"value": 10
//		}
//	]
func (h *handler) UpdateMany(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.showError(w, "cant get body", http.StatusBadRequest)
		return
	}

	if h.Cfg.Settings.Key != "" {
		var metrics []metric.Metric
		err = json.Unmarshal(body, &metrics)
		if err != nil {
			h.showError(w, fmt.Sprintf("cant parse body:%s", string(body)), http.StatusBadRequest)
			return
		}
		for _, v := range metrics {
			if !v.IsValidHash(h.Cfg.Settings.Key) {
				h.showError(w, fmt.Sprintf("wrong hash for %+v", v), http.StatusBadRequest)
				return
			}
		}
	}

	err = h.Storage.ImportFromJSON(r.Context(), body)
	if err != nil {
		h.showError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(nil)
}
