package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuotaTTSJSON_New(t *testing.T) {
	h, err := JSONTTSAsQuota(newTestHandler(), 0.5)
	assert.Nil(t, err)
	assert.NotNil(t, h)
}

func TestNewQuotaTTSJSON_Fail(t *testing.T) {
	_, err := JSONTTSAsQuota(newTestHandler(), -0.5)
	assert.NotNil(t, err)
	_, err = JSONTTSAsQuota(newTestHandler(), 1)
	assert.NotNil(t, err)
	_, err = JSONTTSAsQuota(newTestHandler(), 80)
	assert.NotNil(t, err)
}

func TestQuotaTTSJSON(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Value = "kkk"
	resp := httptest.NewRecorder()
	newJSONTTSHandler(t).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, 3.0, ctx.QuotaValue)
}

func TestQuotaTTSJSON_0(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Value = ""
	resp := httptest.NewRecorder()
	newJSONTTSHandler(t).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, 0.0, ctx.QuotaValue)
}

func TestQuotaTTSJSON_More(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Value = "ą ę ė, olia"
	resp := httptest.NewRecorder()
	newJSONTTSHandler(t).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, 11.0, ctx.QuotaValue)
}

func TestQuotaTTSJSON_Discount(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Value = "ą ę ė, olia"
	b := true
	ctx.Discount = &b
	resp := httptest.NewRecorder()
	newJSONTTSHandler(t).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.InDelta(t, 5.5, ctx.QuotaValue, 0.0001)
}

func TestQuotaTTSJSON_Header(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Value = "ą ę ė, olia"
	ctx.Tags = []string{"aa:oo", allowSaveHeader + ":" + allowSaveValue}
	resp := httptest.NewRecorder()
	newJSONTTSHandler(t).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.InDelta(t, 5.5, ctx.QuotaValue, 0.0001)
}

func newJSONTTSHandler(t *testing.T) http.Handler {
	h, err := JSONTTSAsQuota(newTestHandler(), 0.5)
	assert.Nil(t, err)
	return h
}
