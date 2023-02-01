package metric

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getFloat(t *testing.T) {
	tests := []struct {
		name       string
		input      interface{}
		wantResult float64
	}{
		{
			name:       "test number-string",
			input:      "1.1",
			wantResult: 0,
		},
		{
			name:       "string",
			input:      "test string",
			wantResult: 0,
		},
		{
			name:       "string-number",
			input:      "3.1 ",
			wantResult: 0,
		},
		{
			name:       "int",
			input:      4,
			wantResult: 4,
		},
		{
			name:       "int64",
			input:      int64(123),
			wantResult: 123,
		},
		{
			name:       "float64",
			input:      float64(33.1231),
			wantResult: 33.1231,
		},
		{
			name:       "float32",
			input:      float32(423.1231),
			wantResult: float64(float32(423.1231)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			result, _ := getFloatValue(tt.input)
			require.Equal(result, tt.wantResult, tt.name)
		})
	}
}
