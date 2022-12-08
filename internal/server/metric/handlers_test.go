package metric

import (
	"html/template"
	"net/http"
	"testing"

	"github.com/nickzhog/practicum-metric/pkg/logging"
)

func TestHandler_showError(t *testing.T) {
	type fields struct {
		Data   Storage
		Tpl    *template.Template
		Logger *logging.Logger
	}
	type args struct {
		w      http.ResponseWriter
		err    string
		status int
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
			h := &Handler{
				Data:   tt.fields.Data,
				Tpl:    tt.fields.Tpl,
				Logger: tt.fields.Logger,
			}
			h.showError(tt.args.w, tt.args.err, tt.args.status)
		})
	}
}
