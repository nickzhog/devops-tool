package web

import (
	"context"
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
	"github.com/nickzhog/devops-tool/internal/server/server"
	"github.com/nickzhog/devops-tool/pkg/metric"
)

type handler struct {
	srv server.Server
}

func NewHandler(srv server.Server) *handler {
	return &handler{
		srv: srv,
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
	err := h.srv.Ping(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(nil)
}

// IndexHandler - главная страница, отображает все доступные метрики и их значения
func (h *handler) IndexHandler(w http.ResponseWriter, r *http.Request) {

	metrics, err := h.srv.FindAll(r.Context())
	if err != nil {
		ErrInternalError(err).Render(w, r)
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
		ErrBadRequest(err).Render(w, r)
		return
	}
	var metricElem metric.Metric
	err = json.Unmarshal(body, &metricElem)
	if err != nil {
		ErrUnprocessableEntityRequest(err).Render(w, r)
		return
	}

	metricElem, err = h.srv.FindMetric(r.Context(), metricElem.ID, metricElem.MType)
	if err != nil {
		if err == metric.ErrNoResult {
			ErrNotFound(err).Render(w, r)
			return
		}
		ErrInternalError(err).Render(w, r)
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
		ErrBadRequest(err).Render(w, r)
		return
	}
	var metricElem metric.Metric
	err = json.Unmarshal(body, &metricElem)
	if err != nil {
		ErrUnprocessableEntityRequest(err).Render(w, r)
		return
	}

	h.srv.Logger.Tracef("UpdateFromBody: %s", body)

	err = h.srv.UpsertMetric(r.Context(), metricElem)
	if err != nil {
		ErrInternalError(err).Render(w, r)
		return
	}

	mcurrent, err := h.srv.FindMetric(r.Context(), metricElem.ID, metricElem.MType)
	if err != nil {
		ErrInternalError(err).Render(w, r)
		return
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
		ErrBadRequest(errors.New("metric_type is missing in parameters")).Render(w, r)
		return
	}

	metricName := chi.URLParam(r, "name")
	if len(metricName) < 1 {
		ErrBadRequest(errors.New("name is missing in parameters")).Render(w, r)
		return
	}

	metricElem, err := h.srv.FindMetric(r.Context(), metricName, metricType)
	if err != nil {
		if err == metric.ErrNoResult {
			ErrNotFound(err).Render(w, r)
			return
		}
		ErrInternalError(err).Render(w, r)
		return
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
		ErrBadRequest(errors.New("name is missing in parameters")).Render(w, r)
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
			ErrUnprocessableEntityRequest(err).Render(w, r)
			return
		}
		metricElem = metric.NewGaugeMetric(metricName, value)
		valueString = fmt.Sprintf("%g", value)
	case metric.CounterType:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			ErrUnprocessableEntityRequest(err).Render(w, r)
			return
		}
		metricElem = metric.NewCounterMetric(metricName, value)
		valueString = fmt.Sprintf("%v", value)

		actualMetric, err := h.srv.FindMetric(r.Context(), metricName, metricType)
		if err != nil && !errors.Is(err, metric.ErrNoResult) {
			ErrInternalError(err).Render(w, r)
			return
		} else if err == nil {
			valueString = fmt.Sprintf("%v", *actualMetric.Delta+value)
		}
	default:
		ErrUnprocessableEntityRequest(errors.New("wrong metric type")).Render(w, r)
		return
	}

	err := h.srv.UpsertMetric(r.Context(), metricElem)
	if err != nil {
		ErrInternalError(err).Render(w, r)
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
		ErrBadRequest(err).Render(w, r)
		return
	}

	var metrics []metric.Metric
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		ErrUnprocessableEntityRequest(err).Render(w, r)
		return
	}

	err = h.srv.UpsertMany(r.Context(), metrics)
	if err != nil {
		ErrInternalError(err).Render(w, r)
		return
	}

	w.Write(nil)
}
