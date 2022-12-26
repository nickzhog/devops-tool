package handlers

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/metric"
	"github.com/nickzhog/devops-tool/internal/server/postgresql"
	"github.com/nickzhog/devops-tool/pkg/logging"
)

func (h *Handler) showError(w http.ResponseWriter, err string, status int) {
	// h.Logger.Error(err)
	m := map[string]string{
		"error": err,
	}
	data, _ := json.Marshal(m)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

type Handler struct {
	CacheData metric.Storage
	Logger    *logging.Logger
	Cfg       *config.Config

	// Tpl    *template.Template

	ClientDB    postgresql.Client
	MetricTable metric.Repository
}

func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	_, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()
	err := h.ClientDB.Ping(context.Background())
	if err != nil {
		http.Error(w, fmt.Sprintf("db: %s", err.Error()), http.StatusInternalServerError)
	}
	w.Write(nil)
}

type ForTemplate struct {
	Key   string
	Value interface{}
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

func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	data := h.CacheData.ExportToJSON()

	var metrics []metric.MetricExport
	_ = json.Unmarshal(data, &metrics)

	gaugeData := []ForTemplate{}
	counterData := []ForTemplate{}
	for _, v := range metrics {
		switch v.MType {
		case metric.CounterType:
			counterData = append(counterData, ForTemplate{Key: v.ID, Value: v.Delta})
		case metric.GaugeType:
			gaugeData = append(gaugeData, ForTemplate{Key: v.ID, Value: v.Value})
		}
	}

	m := make(map[string]interface{})

	m["GaugeValues"] = gaugeData

	m["CounterValues"] = counterData

	w.Header().Set("Content-Type", "text/html")
	templ.Execute(w, m)
	// if err := h.Tpl.ExecuteTemplate(w, "index.html", m); err != nil {
	// 	h.showError(w,
	// 		fmt.Sprintf("cant load page:%v", err),
	// 		http.StatusBadGateway)
	// 	return
	// }
}

func (h *Handler) SelectFromBody(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.showError(w, "cant get body", http.StatusBadRequest)
		return
	}
	var metricElem metric.MetricExport
	err = json.Unmarshal(body, &metricElem)
	if err != nil {
		h.showError(w, fmt.Sprintf("cant parse body:%s", string(body)), http.StatusBadRequest)
		return
	}

	var (
		val   interface{}
		exist bool
	)

	switch metricElem.MType {
	case metric.GaugeType:
		val, exist = h.CacheData.FindGaugeByName(metricElem.ID)
	case metric.CounterType:
		val, exist = h.CacheData.FindCounterByName(metricElem.ID)
	default:
		h.showError(w, "wrong metric type", http.StatusBadRequest)
		return
	}
	if !exist {
		h.showError(w, "not found", http.StatusNotFound)
		return
	}

	newMetric := metric.MetricToExport(metricElem.ID, metricElem.MType, val)
	if h.Cfg.Settings.Key != "" {
		newMetric.Hash = string(newMetric.GetHash(h.Cfg.Settings.Key))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(newMetric.Marshal())
}

func (h *Handler) UpdateFromBody(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.showError(w, "cant get body", http.StatusBadRequest)
		return
	}
	var metricElem metric.MetricExport
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

	var (
		newVal interface{}
		ok     bool
	)

	switch metricElem.MType {
	case metric.GaugeType:
		h.CacheData.UpdateGauge(metricElem.ID, *metricElem.Value)
		newVal, ok = h.CacheData.FindGaugeByName(metricElem.ID)
	case metric.CounterType:
		h.CacheData.UpdateCounter(metricElem.ID, *metricElem.Delta)
		newVal, ok = h.CacheData.FindCounterByName(metricElem.ID)
	default:
		h.showError(w, "wrong metric type", http.StatusBadRequest)
		return
	}

	if !ok {
		h.showError(w, "something is wrong", http.StatusBadGateway)
		return
	}

	newMetric := metric.MetricToExport(metricElem.ID, metricElem.MType, newVal)

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
	case metric.GaugeType:
		val, exist = h.CacheData.FindGaugeByName(metricName)
	case metric.CounterType:
		val, exist = h.CacheData.FindCounterByName(metricName)
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

	var (
		newVal interface{}
		ok     bool
	)

	switch metricType {
	case metric.GaugeType:
		metricValueFloat, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			h.showError(w,
				fmt.Sprintf("cant convert value to float:%v", err.Error()),
				http.StatusBadRequest)
			return
		}
		h.CacheData.UpdateGauge(metricName, metricValueFloat)
		newVal, ok = h.CacheData.FindGaugeByName(metricName)
	case metric.CounterType:
		metricValueInt, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			h.showError(w,
				fmt.Sprintf("cant convert value to int64:%v", err.Error()),
				http.StatusBadRequest)
			return
		}
		h.CacheData.UpdateCounter(metricName, metricValueInt)
		newVal, ok = h.CacheData.FindCounterByName(metricName)
	default:
		h.showError(w, "wrong element type", http.StatusNotImplemented)
		return
	}
	if !ok {
		h.showError(w, "something is wrong", http.StatusBadGateway)
		return
	}

	fmt.Fprintf(w, "%v", newVal)
}

func (h *Handler) UpdateMany(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.showError(w, "cant get body", http.StatusBadRequest)
		return
	}
	var metrics []metric.MetricExport
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		h.showError(w, fmt.Sprintf("cant parse body:%s", string(body)), http.StatusBadRequest)
		return
	}

	if h.Cfg.Settings.Key != "" {
		for _, v := range metrics {
			if !hmac.Equal(
				[]byte(v.GetHash(h.Cfg.Settings.Key)),
				[]byte(v.Hash)) {
				h.showError(w, "wrong hash for "+v.ID, http.StatusBadRequest)
				return
			}
		}
	}

	var ok bool

	for _, v := range metrics {
		switch v.MType {
		case metric.GaugeType:
			h.CacheData.UpdateGauge(v.ID, *v.Value)
			_, ok = h.CacheData.FindGaugeByName(v.ID)
		case metric.CounterType:
			h.CacheData.UpdateCounter(v.ID, *v.Delta)
			_, ok = h.CacheData.FindCounterByName(v.ID)
		default:
			h.showError(w, "wrong metric type", http.StatusBadRequest)
			return
		}
		if !ok {
			h.showError(w, "something is wrong", http.StatusBadGateway)
			return
		}
	}

	w.Write(nil)
}
