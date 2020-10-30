package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuotaJSON(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Value = "kkk"
	resp := httptest.NewRecorder()
	JSONAsQuota(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, 3.0, ctx.QuotaValue)
}

func TestQuotaJSON_0(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Value = ""
	resp := httptest.NewRecorder()
	JSONAsQuota(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, 0.0, ctx.QuotaValue)
}

func TestQuotaJSON_More(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Value = "ą ę ė, olia"
	resp := httptest.NewRecorder()
	JSONAsQuota(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, 11.0, ctx.QuotaValue)
}
