package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetric(t *testing.T) {
	delta := int64(10)
	value := float64(10.1)

	type args struct {
		name       string
		metricType string
		value      float64
		delta      int64
	}
	tests := []struct {
		name string
		args args
		want Metric
	}{
		{
			name: "gauge test",
			args: args{
				name:       "good_gauge",
				metricType: GaugeType,
				value:      value,
			},
			want: Metric{
				ID:    "good_gauge",
				MType: GaugeType,
				Value: &value,
			},
		},
		{
			name: "counter test",
			args: args{
				name:       "good_counter",
				metricType: CounterType,
				delta:      delta,
			},
			want: Metric{
				ID:    "good_counter",
				MType: CounterType,
				Delta: &delta,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Metric{}
			switch tt.args.metricType {
			case GaugeType:
				got = NewMetric(tt.args.name, tt.args.metricType, tt.args.value)
			case CounterType:
				got = NewMetric(tt.args.name, tt.args.metricType, tt.args.delta)
			}
			a := assert.New(t)
			a.Equal(tt.want, got)
		})
	}
}
