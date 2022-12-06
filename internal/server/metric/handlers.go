package metric

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func showError(w http.ResponseWriter, err string, status int) {
	fmt.Println("http err:", err)
	http.Error(w, err, status)
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
		gt = append(gt, ForTemplate{Key: k, Value: v})
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

func (h *Handler) SelectHandler(w http.ResponseWriter, r *http.Request) {
	if strings.ToLower(r.Method) != "get" {
		showError(w,
			fmt.Sprintf("method(%v) not get", r.Method),
			http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)

	metricType, ok := vars["metric_type"]
	if !ok {
		showError(w,
			"metric_type is missing in parameters",
			http.StatusBadRequest)
		return
	}

	metricName, ok := vars["name"]
	if !ok {
		showError(w,
			"name is missing in parameters",
			http.StatusBadRequest)
		return
	}

	if metricType == "gauge" {
		v, ok := h.Data.FindGaugeByName(metricName)
		if !ok {
			showError(w, "not found", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "%v", v)
		return

	} else if metricType == "counter" {
		v, ok := h.Data.FindCounterByName(metricName)
		if !ok {
			showError(w, "not found", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "%v", v)
		return
	} else {
		showError(w,
			"wrong element type",
			http.StatusNotImplemented)
		return
	}
}

func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if strings.ToLower(r.Method) != "post" {
		showError(w,
			fmt.Sprintf("method(%v) not post", r.Method),
			http.StatusBadRequest)
		return
	}

	// contentType := r.Header.Get("Content-Type")
	// if contentType != "text/plain" {
	// 	showError(w,
	// 		fmt.Sprintf("Content-Type(%v) not text/plain",
	// 			contentType),
	// 	http.StatusBadRequest)
	// 	return
	// }

	vars := mux.Vars(r)
	metricType, ok := vars["metric_type"]
	if !ok {
		showError(w,
			"metric_type is missing in parameters",
			http.StatusBadRequest)
		return
	}

	metricName, ok := vars["name"]
	if !ok {
		showError(w,
			"name is missing in parameters",
			http.StatusBadRequest)
		return
	}

	metricValue, ok := vars["value"]
	if !ok {
		showError(w,
			"value is missing in parameters",
			http.StatusBadRequest)
		return
	}

	if metricType == "gauge" {
		metricValueFloat, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			showError(w,
				fmt.Sprintf("cant convert value to float:%v", err.Error()),
				http.StatusBadRequest)
			return
		}
		h.Data.UpdateGaugeElem(metricName, metricValueFloat)

	} else if metricType == "counter" {
		metricValueInt, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			showError(w,
				fmt.Sprintf("cant convert value to int64:%v", err.Error()),
				http.StatusBadRequest)
			return
		}

		h.Data.UpdateCounterElem(metricName, metricValueInt)

	} else {
		showError(w,
			"wrong element type",
			http.StatusNotImplemented)
		return
	}

	fmt.Fprint(w, "ok")
}
