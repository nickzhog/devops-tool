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
		Tpl:    nil,
		Logger: nil,
	}

	h.Data.UpdateCounterElem("good_counter", 9)

	type request struct {
		method string
		data   string
	}
	type want struct {
		code        int
		response    string
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
				data:   MetricToJSON("test", GaugeType, 15.1),
			},
			want: want{
				code:        http.StatusOK,
				response:    MetricToJSON("test", GaugeType, 15.1),
				contentType: "application/json",
			},
		},
		{
			name: "wrong type",
			request: request{
				method: http.MethodPost,
				data:   `{"id":"good_metric","type":"new_type", "value": 123}`,
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    string([]byte(`{"error":"wrong metric type"}`)),
				contentType: "application/json",
			},
		},
		{
			name: "existed counter",
			request: request{
				method: http.MethodPost,
				data:   MetricToJSON("good_counter", CounterType, 10),
			},
			want: want{
				code:        http.StatusOK,
				response:    MetricToJSON("good_counter", CounterType, 19),
				contentType: "application/json",
			},
		},
		{
			name: "empty body",
			request: request{
				method: http.MethodPost,
				data:   ``,
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    string([]byte(`{"error":"cant parse body:"}`)),
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

			assert.Equal(tt.want.response, string(resBody))
			assert.Equal(tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
