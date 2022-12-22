package metric

import (
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/nickzhog/practicum-metric/internal/server/config"
	"github.com/nickzhog/practicum-metric/pkg/logging"
)

func (h *Handler) showError(w http.ResponseWriter, err string, status int) {
	// h.Logger.Error(err)
	m := map[string]string{
		"error": err,
	}
	errJSON, _ := json.Marshal(m)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprint(w, string(errJSON))
}

type Handler struct {
	Data   Storage
	Tpl    *template.Template
	Logger *logging.Logger
	Cfg    *config.Config
}

type ForTemplate struct {
	Key   string
	Value interface{}
}

func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	data := h.Data.ExportToJSON()

	var metrics []MetricExport
	_ = json.Unmarshal(data, &metrics)

	gaugeData := []ForTemplate{}
	counterData := []ForTemplate{}
	for _, v := range metrics {
		switch v.MType {
		case CounterType:
			counterData = append(counterData, ForTemplate{Key: v.ID, Value: v.Delta})
		case GaugeType:
			gaugeData = append(gaugeData, ForTemplate{Key: v.ID, Value: v.Value})
		}
	}

	m := make(map[string]interface{})

	m["GaugeValues"] = gaugeData

	m["CounterValues"] = counterData

	if err := h.Tpl.ExecuteTemplate(w, "index.html", m); err != nil {
		h.showError(w,
			fmt.Sprintf("cant load page:%v", err),
			http.StatusBadGateway)
	}
}

func (h *Handler) SelectFromBody(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.showError(w, "cant get body", http.StatusBadRequest)
		return
	}
	var metric MetricExport
	err = json.Unmarshal(body, &metric)
	if err != nil {
		h.showError(w, fmt.Sprintf("cant parse body:%s", string(body)), http.StatusBadRequest)
		return
	}

	var (
		val   interface{}
		exist bool
	)

	switch metric.MType {
	case GaugeType:
		val, exist = h.Data.FindGaugeByName(metric.ID)
	case CounterType:
		val, exist = h.Data.FindCounterByName(metric.ID)
	default:
		h.showError(w, "wrong metric type", http.StatusBadRequest)
		return
	}
	if !exist {
		h.showError(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(MetricToExport(metric.ID, metric.MType, val).Marshal())
}

func (h *Handler) UpdateFromBody(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.showError(w, "cant get body", http.StatusBadRequest)
		return
	}
	var metric MetricExport
	err = json.Unmarshal(body, &metric)
	if err != nil {
		h.showError(w, fmt.Sprintf("cant parse body:%s", string(body)), http.StatusBadRequest)
		return
	}

	if h.Cfg.Settings.Key != "" {
		if !hmac.Equal(
			[]byte(metric.GetHash(h.Cfg.Settings.Key)),
			[]byte(metric.Hash)) {
			h.showError(w, "wrong hash", http.StatusBadRequest)
			return
		}
	}

	var (
		newVal interface{}
		ok     bool
	)

	switch metric.MType {
	case GaugeType:
		h.Data.UpdateGauge(metric.ID, *metric.Value)
		newVal, ok = h.Data.FindGaugeByName(metric.ID)
	case CounterType:
		h.Data.UpdateCounter(metric.ID, *metric.Delta)
		newVal, ok = h.Data.FindCounterByName(metric.ID)
	default:
		h.showError(w, "wrong metric type", http.StatusBadRequest)
		return
	}
	if !ok {
		h.showError(w, "something is wrong", http.StatusBadGateway)
		return
	}
	newMetric := MetricToExport(metric.ID, metric.MType, newVal)
	if h.Cfg.Settings.Key != "" {
		newMetric.Hash = string(newMetric.GetHash(h.Cfg.Settings.Key))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(newMetric.Marshal())
}

func (h *Handler) SelectFromURL(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metric_type")
	if len(metricType) < 1 {
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

	var (
		val   interface{}
		exist bool
	)

	switch metricType {
	case GaugeType:
		val, exist = h.Data.FindGaugeByName(metricName)
	case CounterType:
		val, exist = h.Data.FindCounterByName(metricName)
	default:
		h.showError(w, "wrong metric type", http.StatusBadRequest)
		return
	}
	if !exist {
		h.showError(w, "not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%v", val)
}

func (h *Handler) UpdateFromURL(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metric_type")
	if len(metricType) < 1 {
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

	metricValue := chi.URLParam(r, "value")

	switch metricType {
	case GaugeType:
		metricValueFloat, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			h.showError(w,
				fmt.Sprintf("cant convert value to float:%v", err.Error()),
				http.StatusBadRequest)
			return
		}
		h.Data.UpdateGauge(metricName, metricValueFloat)
	case CounterType:
		metricValueInt, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			h.showError(w,
				fmt.Sprintf("cant convert value to int64:%v", err.Error()),
				http.StatusBadRequest)
			return
		}
		h.Data.UpdateCounter(metricName, metricValueInt)
	default:
		h.showError(w,
			"wrong element type",
			http.StatusNotImplemented)
		return
	}

	fmt.Fprintf(w, "%s", metricValue)
}
