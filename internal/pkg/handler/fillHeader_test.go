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

func TestFillKeyHeader(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "olia"
	resp := httptest.NewRecorder()
	FillKeyHeader(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "key_72bee22eaaf36e984ff30033298c5932", req.Header.Get(headerSaveTags))
}

func TestFillKeyHeader_NoKey(t *testing.T) {
	req, _ := customContext(httptest.NewRequest("POST", "/duration", nil))
	resp := httptest.NewRecorder()
	FillKeyHeader(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "", req.Header.Get(headerSaveTags))
}

func TestHash(t *testing.T) {
	assert.Equal(t, "0cc175b9c0f1b6a831c399e269772661", hashKey("a"))
	assert.Equal(t, "d2392c572c25241d3f042ec06d2ed990", hashKey("loooooooooooooooooooooooooooooooooooooooooooooong"))
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
