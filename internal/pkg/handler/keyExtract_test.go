package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyExtract(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration?key=oooo", strings.NewReader(`{"body":"olia"}`)))
	resp := httptest.NewRecorder()

	KeyExtract(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "oooo", ctx.Key)
	assert.True(t, ctx.Manual)
}

func TestKeyExtract_Empty(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", strings.NewReader(`{"body":"olia"}`)))
	resp := httptest.NewRecorder()

	KeyExtract(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, "", ctx.Key)
	assert.False(t, ctx.Manual)
	assert.Equal(t, testCode, resp.Code)
}

func TestKeyExtract_FromHeader(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest(http.MethodPost, "/duration", strings.NewReader(`{"body":"olia"}`)))
	req.Header.Set(authHeader, "Key olia")
	resp := httptest.NewRecorder()

	KeyExtract(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, "olia", ctx.Key)
	assert.True(t, ctx.Manual)
	assert.Equal(t, "", req.Header.Get(authHeader))
	assert.Equal(t, testCode, resp.Code)
}

func TestKeyExtract_TrimParam(t *testing.T) {
	req, _ := customContext(httptest.NewRequest("POST", "/duration?key=oooo&key1=111", strings.NewReader(`{"body":"olia}`)))
	resp := httptest.NewRecorder()

	KeyExtract(newTestHandler()).ServeHTTP(resp, req)

	q, _ := url.ParseQuery(req.URL.RawQuery)

	key := q.Get("key")
	assert.Equal(t, "", key)
	key1 := q.Get("key1")
	assert.Equal(t, "111", key1)
}

func TestKeyExtract_HeaderFirst(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest(http.MethodPost, "/duration?key=oooo", strings.NewReader(`{"body":"olia}`)))
	resp := httptest.NewRecorder()
	req.Header.Set(authHeader, "Key olia")

	KeyExtract(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, "olia", ctx.Key)
	assert.True(t, ctx.Manual)
	assert.Equal(t, testCode, resp.Code)
}

func TestKeyExtract_FailHeader(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest(http.MethodPost, "/duration",
		strings.NewReader(`{"body":"olia}`)))
	resp := httptest.NewRecorder()
	req.Header.Set(authHeader, "Key xx xx")

	KeyExtract(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, "", ctx.Key)
	assert.False(t, ctx.Manual)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func Test_extractKey(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "Empty", args: args{str: ""}, want: "", wantErr: false},
		{name: "Extracts", args: args{str: "Key olia"}, want: "olia", wantErr: false},
		{name: "Fails", args: args{str: "Key"}, want: "", wantErr: true},
		{name: "Fails", args: args{str: "olia"}, want: "", wantErr: true},
		{name: "Fails", args: args{str: "Key olia olia"}, want: "", wantErr: true},
		{name: "Fails", args: args{str: " "}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractKey(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
