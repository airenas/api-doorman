package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTTSJSON(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", strings.NewReader(`{"text":"olia"}`)))
	resp := httptest.NewRecorder()

	TakeJSONTTS(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "olia", ctx.Value)
	assert.Nil(t, ctx.Discount)
}

func TestTTSJSON_Discount(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration",
		strings.NewReader(`{"text":"olia", "allowCollectData":true}`)))
	resp := httptest.NewRecorder()

	TakeJSONTTS(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "olia", ctx.Value)
	assert.Equal(t, true, *ctx.Discount)
}

func TestTTSJSON_Empty(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", strings.NewReader(`{"text":""}`)))
	resp := httptest.NewRecorder()

	TakeJSONTTS(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, "", ctx.Value)
	assert.Equal(t, testCode, resp.Code)
}

func TestTTSJSON_Fail(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", strings.NewReader(`{"text":"olia}`)))
	resp := httptest.NewRecorder()

	TakeJSONTTS(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, "", ctx.Value)
	assert.Equal(t, 400, resp.Code)
}

func TestTTSJSON_FailNotString(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", strings.NewReader(`{"text":10}`)))
	resp := httptest.NewRecorder()

	TakeJSONTTS(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, "", ctx.Value)
	assert.Equal(t, 400, resp.Code)
}

func TestTTSJSON_Parses(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", strings.NewReader(`{"text":"10", "opa": 20, "hi":true,"a":["aa"]}`)))
	resp := httptest.NewRecorder()

	TakeJSONTTS(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, "10", ctx.Value)
	assert.Equal(t, testCode, resp.Code)
}
