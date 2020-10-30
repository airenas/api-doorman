package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogStdout(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.QuotaValue = 100
	ctx.Value = "olia"
	resp := httptest.NewRecorder()
	h := LogStdout(newTestHandler()).(*logStdout)
	sb := &strings.Builder{}
	h.log = sb

	h.ServeHTTP(resp, req)

	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "192.0.2.1 100.00 'olia'", sb.String())
}
