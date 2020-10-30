package handler

import (
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
