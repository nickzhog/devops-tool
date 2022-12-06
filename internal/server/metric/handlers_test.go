package metric

import (
	"net/http"
	"testing"
)

func Test_showError(t *testing.T) {
	type args struct {
		w      http.ResponseWriter
		err    string
		status int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			showError(tt.args.w, tt.args.err, tt.args.status)
		})
	}
}
