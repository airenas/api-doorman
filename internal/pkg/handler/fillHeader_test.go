package handler

import (
	"net/http"
	"net/http/httptest"
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

func TestHeaderValue(t *testing.T) {
	h, v, err := headerValue("")
	assert.Nil(t, err)
	assert.Equal(t, "", h)
	assert.Equal(t, "", v)
	h, v, err = headerValue("aaa:oooo")
	assert.Nil(t, err)
	assert.Equal(t, "aaa", h)
	assert.Equal(t, "oooo", v)
	h, v, err = headerValue(":aaa:oooo")
	assert.Nil(t, err)
	assert.Equal(t, "", h)
	assert.Equal(t, "aaa:oooo", v)
	h, v, err = headerValue("aaaa:")
	assert.Nil(t, err)
	assert.Equal(t, "aaaa", h)
	assert.Equal(t, "", v)
}

func TestHeaderValue_Fail(t *testing.T) {
	h, v, err := headerValue("asdasd")
	assert.NotNil(t, err)
	assert.Equal(t, "", h)
	assert.Equal(t, "", v)
}
