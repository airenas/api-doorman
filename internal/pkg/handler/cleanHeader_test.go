package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanHeader(t *testing.T) {
	type args struct {
		next     http.Handler
		starting string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantHead []string
	}{
		{name: "OK", args: args{next: nil, starting: "st"}, wantErr: false},
		{name: "Fail empty", args: args{next: nil, starting: ""}, wantErr: true},
		{name: "Fail", args: args{next: nil, starting: " "}, wantErr: true},
		{name: "Several", args: args{next: nil, starting: "H1, h2"}, wantErr: false, wantHead: []string{"H1", "H2"}},
		{name: "Several", args: args{next: nil, starting: ",,, H1, h2,,x-H3"}, wantErr: false, wantHead: []string{"H1", "H2", "X-H3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CleanHeader(tt.args.next, tt.args.starting)
			if (err != nil) != tt.wantErr {
				t.Errorf("CleanHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.NotNil(t, got)
			}
			if len(tt.wantHead) > 0 {
				cgot, _ := got.(*cleanHeader)
				if assert.Equal(t, len(tt.wantHead), len(cgot.headers)) {
					assert.Equal(t, tt.wantHead, cgot.headers)
				}
			}
		})
	}
}

func Test_cleanHeader_ServeHTTP(t *testing.T) {
	tests := []struct {
		name string
		pr   string
		in   map[string][]string
		out  http.Header
	}{
		{name: "Leaves", pr: "none", in: map[string][]string{"H1": {"olia"}},
			out: map[string][]string{"H1": {"olia"}}},
		{name: "Drops", pr: "x-prefix", in: map[string][]string{"H1": {"olia"}, "x-prefix": {"aaa"}},
			out: map[string][]string{"H1": {"olia"}}},
		{name: "Drops", pr: "x-prefix",
			in: map[string][]string{"H1": {"olia"}, "x-prefix": {"aaa"},
				"x-prefix-2": {"aaa", "bbb"}},
			out: map[string][]string{"H1": {"olia"}, "X-Prefix-2": {"aaa", "bbb"}}},
		{name: "Leaves", pr: "x-prefix-",
			in: map[string][]string{"H1": {"olia", "aaa"}, "x-prefix": {"aaa"},
				"x-prefix-2": {"aaa", "bbb"}},
			out: map[string][]string{"H1": {"olia", "aaa"}, "X-Prefix": {"aaa"}, "X-Prefix-2": {"aaa", "bbb"}}},
		{name: "Drops two", pr: "x-prefix,x-prefix-2",
			in: map[string][]string{"H1": {"olia", "aaa"}, "x-prefix": {"aaa"},
				"X-Prefix-2": {"aaa", "bbb"}},
			out: map[string][]string{"H1": {"olia", "aaa"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := customContext(httptest.NewRequest("POST", "/duration", nil))
			for k, v := range tt.in {
				for _, s := range v {
					req.Header.Add(k, s)
				}
			}
			resp := httptest.NewRecorder()
			h, _ := CleanHeader(newTestHandler(), tt.pr)
			h.ServeHTTP(resp, req)
			assert.Equal(t, tt.out, req.Header)
			assert.Equal(t, "", req.Header.Get(headerSaveTags))
		})
	}
}
