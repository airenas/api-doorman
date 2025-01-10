package utils

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

type td struct {
	A string
}

func TestTakeJSONInput(t *testing.T) {
	type args struct {
		inType string
		in     io.Reader
	}
	tests := []struct {
		name    string
		args    args
		out     interface{}
		wantErr bool
	}{
		{name: "OK", args: args{inType: echo.MIMEApplicationJSON, in: strings.NewReader(`{"a":"b"}`)},
			out: &td{}, wantErr: false},
		{name: "Content Type", args: args{inType: echo.MIMEApplicationXMLCharsetUTF8,
			in: strings.NewReader(`{"a":"b"}`)},
			out: &td{}, wantErr: true},
		{name: "Content Type", args: args{inType: "",
			in: strings.NewReader(`{"a":"b"}`)},
			out: &td{}, wantErr: true},
		{name: "Input", args: args{inType: echo.MIMEApplicationJSON, in: strings.NewReader(`aaa{"a":"b"}`)},
			out: &td{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", tt.args.in)
			req.Header.Add(echo.HeaderContentType, tt.args.inType)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := TakeJSONInput(c, tt.out); (err != nil) != tt.wantErr {
				t.Errorf("TakeJSONInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
