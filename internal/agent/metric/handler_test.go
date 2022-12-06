package metric

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_getFloat(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "test number-string",
			input: "1.1",
		},
		{
			name:  "string",
			input: "test string",
		},
		{
			name:  "string-number",
			input: "3.1 ",
		},
		{
			name:  "int",
			input: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			result, ok := getFloat(tt.input)
			fmt.Printf("getFloat() input:%v,type:%v, result:%v, ok:%v\n",
				tt.input, reflect.TypeOf(tt.input), result, ok)
		})
	}
}
