package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFillHeader(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Tags = []string{"olia:15"}
	resp := httptest.NewRecorder()
	FillHeader(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "15", req.Header.Get("olia"))
}

func TestFillHeader_Several(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Tags = []string{"olia:15", "xkkk: 12:16"}
	resp := httptest.NewRecorder()
	FillHeader(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "15", req.Header.Get("olia"))
	assert.Equal(t, "12:16", req.Header.Get("xkkk"))
}

func TestFillHeader_Fail(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Tags = []string{"olia=15"}
	resp := httptest.NewRecorder()
	FillHeader(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestFillKeyHeader(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "olia"
	resp := httptest.NewRecorder()
	FillKeyHeader(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "key_d57329cf35", req.Header.Get(headerSaveTags))
}

func TestFillKeyHeader_NoKey(t *testing.T) {
	req, _ := customContext(httptest.NewRequest("POST", "/duration", nil))
	resp := httptest.NewRecorder()
	FillKeyHeader(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "", req.Header.Get(headerSaveTags))
}

func TestHash(t *testing.T) {
	assert.Equal(t, "ca978112ca", hashKey("a"))
	assert.Equal(t, "168a05449e", hashKey("loooooooooooooooooooooooooooooooooooooooooooooong"))
	assert.Equal(t, "82396ec919", hashKey(strings.Repeat("aaaa", 1000)))
}

func TestSetHeader(t *testing.T) {
	req := httptest.NewRequest("POST", "/duration", nil)
	setHeader(req, headerSaveTags, "olia")
	assert.Equal(t, "olia", req.Header.Get(headerSaveTags))
	setHeader(req, headerSaveTags, "olia2")
	assert.Equal(t, "olia,olia2", req.Header.Get(headerSaveTags))
	setHeader(req, headerSaveTags, "")
	assert.Equal(t, "olia,olia2", req.Header.Get(headerSaveTags))
}

func Test_trim(t *testing.T) {
	type args struct {
		s string
		i int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Empty", args: args{s: "", i: 10}, want: ""},
		{name: "Full", args: args{s: "aaa", i: 10}, want: "aaa"},
		{name: "Full exact", args: args{s: "aaaaa", i: 5}, want: "aaaaa"},
		{name: "Trim", args: args{s: "aaaaaaa", i: 5}, want: "aaaaa"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trim(tt.args.s, tt.args.i); got != tt.want {
				t.Errorf("trim() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFillRequestIDHeader(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.RequestID = "oliaID"
	resp := httptest.NewRecorder()
	FillRequestIDHeader(newTestHandler(), "testdb").ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "testdb::oliaID", req.Header.Get(headerRequestID))
}

func TestFillRequestIDHeader_Manual(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.RequestID = "oliaID"
	ctx.Manual = true
	resp := httptest.NewRecorder()
	FillRequestIDHeader(newTestHandler(), "testdb").ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "testdb:m:oliaID", req.Header.Get(headerRequestID))
}

func TestFillRequestIDHeader_NoHeader(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.RequestID = ""
	resp := httptest.NewRecorder()
	FillRequestIDHeader(newTestHandler(), "testdb").ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "", req.Header.Get(headerRequestID))
}
