package metric

import "testing"

func TestMetrics_InitMetrics(t *testing.T) {
	type fields struct {
		GaugeMetrics   map[string]float64
		CounterMetrics map[string]int64
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				GaugeMetrics:   tt.fields.GaugeMetrics,
				CounterMetrics: tt.fields.CounterMetrics,
			}
			m.InitMetrics()
		})
	}
}
