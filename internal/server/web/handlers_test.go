package web

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/metric"
	"github.com/nickzhog/devops-tool/internal/server/metric/cache"
	"github.com/nickzhog/devops-tool/pkg/logging"
	"github.com/stretchr/testify/assert"
)

func TestHandler_UpdateFromBody(t *testing.T) {
	h := NewHandlerData(logging.GetLogger(), &config.Config{}, cache.NewMemStorage())

	h.Cfg.Settings.Key = ""

	type request struct {
		data []byte
	}
	type want struct {
		code        int
		response    []byte
		contentType string
	}
	tests := []struct {
		name string
		request
		want
	}{
		{
			name: "valid gauge metric",
			request: request{
				data: metric.NewGaugeMetric("test_gauge", 15.1).Marshal(),
			},
			want: want{
				code:        http.StatusOK,
				response:    metric.NewGaugeMetric("test_gauge", 15.1).Marshal(),
				contentType: "application/json",
			},
		},
		{
			name: "wrong type",
			request: request{
				data: []byte(`{"id":"good_metric","type":"new_type", "value": 123}`),
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    []byte(`{"error":"wrong metric type"}`),
				contentType: "application/json",
			},
		},
		{
			name: "valid counter",
			request: request{
				data: metric.NewCounterMetric("good_counter", 10).Marshal(),
			},
			want: want{
				code:        http.StatusOK,
				response:    metric.NewCounterMetric("good_counter", 10).Marshal(),
				contentType: "application/json",
			},
		},
		{
			name: "counter increment",
			request: request{
				data: metric.NewCounterMetric("good_counter", 1).Marshal(),
			},
			want: want{
				code:        http.StatusOK,
				response:    metric.NewCounterMetric("good_counter", 11).Marshal(),
				contentType: "application/json",
			},
		},
		{
			name: "empty body",
			request: request{
				data: nil,
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    []byte(`{"error":"cant parse body:"}`),
				contentType: "application/json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer([]byte(tt.data)))

			w := httptest.NewRecorder()
			h := http.HandlerFunc(h.UpdateFromBody)
			h.ServeHTTP(w, request)
			res := w.Result()

			assert := assert.New(t)
			assert.Equal(tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			assert.NoError(err)

			assert.JSONEq(string(tt.want.response), string(resBody))
			assert.Equal(tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestHandler_SelectFromBody(t *testing.T) {
	h := NewHandlerData(logging.GetLogger(), &config.Config{}, cache.NewMemStorage())
	h.Cfg.Settings.Key = ""

	nulFloat := float64(0)
	nulInt := int64(0)

	type want struct {
		code     int
		response []byte
	}
	tests := []struct {
		name   string
		metric metric.Metric
		want
	}{
		{
			name:   "gauge test",
			metric: metric.NewGaugeMetric("good_gauge", 10),
			want: want{
				code:     http.StatusOK,
				response: []byte(`{"id":"good_gauge","type":"gauge", "value": 10}`),
			},
		},
		{
			name:   "counter test",
			metric: metric.NewCounterMetric("good_counter", 10),
			want: want{
				code:     http.StatusOK,
				response: []byte(`{"id":"good_counter","type":"counter", "delta": 10}`),
			},
		},
		{
			name:   "counter increment",
			metric: metric.NewCounterMetric("good_counter", 10),
			want: want{
				code:     http.StatusOK,
				response: []byte(`{"id":"good_counter","type":"counter", "delta": 20}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			err := h.Storage.UpsertMetric(context.Background(), tt.metric)
			assert.NoError(err)

			tt.metric.Value = &nulFloat
			tt.metric.Delta = &nulInt
			request := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(tt.metric.Marshal()))

			w := httptest.NewRecorder()
			h := http.HandlerFunc(h.SelectFromBody)
			h.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			assert.NoError(err)

			assert.JSONEq(string(tt.want.response), string(resBody))
		})
	}
}

func TestHandler_UpdateMany(t *testing.T) {
	handler := NewHandlerData(logging.GetLogger(), &config.Config{}, cache.NewMemStorage())
	handler.Cfg.Settings.Key = ""

	tests := []struct {
		name        string
		requestData []byte
		wantCode    int
	}{
		{
			name: "gauge metrics",
			requestData: []byte(`
			[
				{"id":"good_metric","type":"gauge", "value": 321},
				{"id":"good_metric","type":"gauge", "value": 123}
			]
			`),
			wantCode: http.StatusOK,
		},
		{
			name: "counter increment",
			requestData: []byte(`
			[
				{"id":"good_metric","type":"gauge","value":321},
				{"id":"good_metric","type":"gauge","value":123},
				{"id":"good_metric","type":"counter","delta":10},
				{"id":"good_metric","type":"counter","delta":10}
			]
			`),
			wantCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler.Storage = cache.NewMemStorage()

			request := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer([]byte(tt.requestData)))

			w := httptest.NewRecorder()
			h := http.HandlerFunc(handler.UpdateMany)
			h.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			assert := assert.New(t)
			assert.Equal(tt.wantCode, res.StatusCode)

			if tt.wantCode == http.StatusOK {
				tempStorage := cache.NewMemStorage()
				err := tempStorage.ImportFromJSON(context.TODO(), tt.requestData)
				assert.NoError(err)

				data, err := tempStorage.ExportToJSON(context.TODO())
				assert.NoError(err)

				dataHandler, err := handler.Storage.ExportToJSON(context.TODO())
				assert.NoError(err)

				assert.JSONEq(string(data), string(dataHandler))
			}
		})
	}
}

func TestHandler_UpdateFromURL(t *testing.T) {
	handler := NewHandlerData(logging.GetLogger(), &config.Config{}, cache.NewMemStorage())

	handler.Cfg.Settings.Key = ""

	type want struct {
		code     int
		response []byte
	}
	tests := []struct {
		name   string
		metric metric.Metric
		want
	}{
		{
			name:   "gauge test",
			metric: metric.NewGaugeMetric("good_gauge", 8.2),
			want: want{
				code:     http.StatusOK,
				response: []byte(`8.2`),
			},
		},
		{
			name:   "counter test",
			metric: metric.NewCounterMetric("good_counter", 10),
			want: want{
				code:     http.StatusOK,
				response: []byte(`10`),
			},
		},
		{
			name:   "counter increment",
			metric: metric.NewCounterMetric("good_counter", 10),
			want: want{
				code:     http.StatusOK,
				response: []byte(`20`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/{metric_type}/{name}/{value}", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("metric_type", tt.metric.MType)
			rctx.URLParams.Add("name", tt.metric.ID)

			switch tt.metric.MType {
			case metric.GaugeType:
				rctx.URLParams.Add("value", fmt.Sprintf("%g", *tt.metric.Value))

			case metric.CounterType:
				rctx.URLParams.Add("value", fmt.Sprintf("%v", *tt.metric.Delta))
			}

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h := http.HandlerFunc(handler.UpdateFromURL)
			h.ServeHTTP(w, r)
			res := w.Result()

			assert.Equal(tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			assert.NoError(err)

			assert.Equal(string(tt.want.response), string(resBody))
		})
	}
}

func TestHandler_SelectFromURL(t *testing.T) {
	handler := NewHandlerData(logging.GetLogger(), &config.Config{}, cache.NewMemStorage())
	handler.Cfg.Settings.Key = ""

	type want struct {
		code     int
		response []byte
	}
	tests := []struct {
		name   string
		metric metric.Metric
		want
	}{
		{
			name:   "gauge test",
			metric: metric.NewGaugeMetric("good_gauge", 10),
			want: want{
				code:     http.StatusOK,
				response: []byte(`10`),
			},
		},
		{
			name:   "counter test",
			metric: metric.NewCounterMetric("good_counter", 10),
			want: want{
				code:     http.StatusOK,
				response: []byte(`10`),
			},
		},
		{
			name:   "counter increment",
			metric: metric.NewCounterMetric("good_counter", 10),
			want: want{
				code:     http.StatusOK,
				response: []byte(`20`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			err := handler.Storage.UpsertMetric(context.Background(), tt.metric)
			assert.NoError(err)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/{metric_type}/{name}", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("metric_type", tt.metric.MType)
			rctx.URLParams.Add("name", tt.metric.ID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			h := http.HandlerFunc(handler.SelectFromURL)
			h.ServeHTTP(w, r)
			res := w.Result()

			assert.Equal(tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			assert.NoError(err)

			assert.Equal(string(tt.want.response), string(resBody))
		})
	}
}
