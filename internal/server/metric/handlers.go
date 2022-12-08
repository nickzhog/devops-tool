package metric

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func showError(w http.ResponseWriter, err string, status int) {
	m := map[string]interface{}{
		"error": err,
	}
	errJson, _ := json.Marshal(m)
	http.Error(w, string(errJson), status)
}

type Handler struct {
	Data Storage
	Tpl  *template.Template
}

type ForTemplate struct {
	Key   string
	Value interface{}
}

func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	memStorage := h.Data.FindAll()

	gt := []ForTemplate{}
	for k, v := range memStorage.GaugeMetrics {
		gt = append(gt, ForTemplate{Key: k, Value: fmt.Sprintf("%f", v)})
	}

	ct := []ForTemplate{}
	for k, v := range memStorage.CounterMetrics {
		ct = append(ct, ForTemplate{Key: k, Value: v})
	}

	m := make(map[string]interface{})

	m["GaugeValues"] = gt

	m["CounterValues"] = ct

	if err := h.Tpl.ExecuteTemplate(w, "index.html", m); err != nil {
		showError(w,
			fmt.Sprintf("cant load page:%v", err),
			http.StatusBadGateway)
		return
	}
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func metricsToJSON(name, metricType string, value interface{}) string {
	m := make(map[string]interface{})
	m["name"] = name
	m["type"] = metricType
	switch metricType {
	case gaugeType:
		m["delta"] = value
	case counterType:
		m["value"] = value
	}

	ans, err := json.Marshal(m)
	if err != nil {
		log.Fatalf("metricsToJSON: %s", err.Error())
	}

	return string(ans)
}

func (h *Handler) SelectFromBody(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		showError(w, "cant get body", http.StatusBadRequest)
		return
	}
	var metric Metrics
	err = json.Unmarshal(body, &metric)
	if err != nil {
		showError(w, "cant parse body", http.StatusBadRequest)
		return
	}

	var (
		val   interface{}
		exist bool
	)

	switch metric.MType {
	case gaugeType:
		val, exist = h.Data.FindGaugeByName(metric.ID)
	case counterType:
		val, exist = h.Data.FindCounterByName(metric.ID)
	default:
		showError(w, "wrong metric type", http.StatusBadRequest)
		return
	}
	if !exist {
		showError(w, "not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%s", metricsToJSON(metric.ID, metric.MType, val))
}

func (h *Handler) UpdateFromBody(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		showError(w, "cant get body", http.StatusBadRequest)
		return
	}
	var metric Metrics
	err = json.Unmarshal(body, &metric)
	if err != nil {
		showError(w, "cant parse body", http.StatusBadRequest)
		return
	}

	var (
		newVal interface{}
		ok     bool
	)

	switch metric.MType {
	case gaugeType:
		h.Data.UpdateGaugeElem(metric.ID, *metric.Value)
		newVal, ok = h.Data.FindGaugeByName(metric.ID)
	case counterType:
		h.Data.UpdateCounterElem(metric.ID, *metric.Delta)
		newVal, ok = h.Data.FindCounterByName(metric.ID)
	default:
		showError(w, "wrong metric type", http.StatusBadRequest)
	}
	if !ok {
		showError(w, "something is wrong", http.StatusBadGateway)
		return
	}

	fmt.Fprintf(w, "%s", metricsToJSON(metric.ID, metric.MType, newVal))
}

func (h *Handler) SelectHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metric_type")
	if len(metricType) < 1 {
		showError(w,
			"metric_type is missing in parameters",
			http.StatusBadRequest)
		return
	}

	metricName := chi.URLParam(r, "name")
	if len(metricName) < 1 {
		showError(w,
			"name is missing in parameters",
			http.StatusBadRequest)
		return
	}

	var (
		val   interface{}
		exist bool
	)

	switch metricType {
	case gaugeType:
		val, exist = h.Data.FindGaugeByName(metricName)
	case counterType:
		val, exist = h.Data.FindCounterByName(metricName)
	default:
		showError(w, "wrong metric type", http.StatusBadRequest)
		return
	}
	if !exist {
		showError(w, "not found", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "%s", metricsToJSON(metricName, metricType, val))
}

func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metric_type")
	if len(metricType) < 1 {
		showError(w,
			"metric_type is missing in parameters",
			http.StatusBadRequest)
		return
	}

	metricName := chi.URLParam(r, "name")
	if len(metricName) < 1 {
		showError(w,
			"name is missing in parameters",
			http.StatusBadRequest)
		return
	}

	metricValue := chi.URLParam(r, "value")

	switch metricType {
	case gaugeType:
		metricValueFloat, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			showError(w,
				fmt.Sprintf("cant convert value to float:%v", err.Error()),
				http.StatusBadRequest)
			return
		}
		h.Data.UpdateGaugeElem(metricName, metricValueFloat)
	case counterType:
		metricValueInt, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			showError(w,
				fmt.Sprintf("cant convert value to int64:%v", err.Error()),
				http.StatusBadRequest)
			return
		}
		h.Data.UpdateCounterElem(metricName, metricValueInt)
	default:
		showError(w,
			"wrong element type",
			http.StatusNotImplemented)
		return
	}

	fmt.Fprintf(w, "%s", metricsToJSON(metricName, metricType, metricValue))
}
