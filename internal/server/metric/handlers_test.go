package metric

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_UpdateFromBody(t *testing.T) {
	h := &Handler{
		Data:   NewMemStorage(),
		Logger: nil,
	}

	h.Data.UpdateCounter("good_counter", 9)

	type request struct {
		method string
		data   []byte
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
			name: "positive case #1",
			request: request{
				method: http.MethodPost,
				data:   MetricToExport("test", GaugeType, float64(15.1)).Marshal(),
			},
			want: want{
				code:        http.StatusOK,
				response:    MetricToExport("test", GaugeType, float64(15.1)).Marshal(),
				contentType: "application/json",
			},
		},
		{
			name: "wrong type",
			request: request{
				method: http.MethodPost,
				data:   []byte(`{"id":"good_metric","type":"new_type", "value": 123}`),
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    []byte(`{"error":"wrong metric type"}`),
				contentType: "application/json",
			},
		},
		{
			name: "existed counter",
			request: request{
				method: http.MethodPost,
				data:   MetricToExport("good_counter", CounterType, int64(10)).Marshal(),
			},
			want: want{
				code:        http.StatusOK,
				response:    MetricToExport("good_counter", CounterType, int64(19)).Marshal(),
				contentType: "application/json",
			},
		},
		{
			name: "empty body",
			request: request{
				method: http.MethodPost,
				data:   []byte(``),
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
			request := httptest.NewRequest(tt.method, "/update", bytes.NewBuffer([]byte(tt.data)))

			w := httptest.NewRecorder()
			h := http.HandlerFunc(h.UpdateFromBody)
			h.ServeHTTP(w, request)
			res := w.Result()

			assert := assert.New(t)
			assert.Equal(tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			assert.NoError(err)

			assert.Equal(string(tt.want.response), string(resBody))
			assert.Equal(tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
