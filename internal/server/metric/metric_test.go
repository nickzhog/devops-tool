package metric

import "testing"

func TestMemStorage_UpdateGaugeElem(t *testing.T) {
	type fields struct {
		GaugeMetrics   map[string]float64
		CounterMetrics map[string]int64
	}
	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				GaugeMetrics:   tt.fields.GaugeMetrics,
				CounterMetrics: tt.fields.CounterMetrics,
			}
			m.UpdateGaugeElem(tt.args.name, tt.args.value)
		})
	}
}
