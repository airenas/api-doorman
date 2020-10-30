package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequoestQuota(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	resp := httptest.NewRecorder()
	RequestAsQuota(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, 1.0, ctx.QuotaValue)
}
