package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyOrIP_Key(t *testing.T) {
	initKeyValidatorTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	resp := httptest.NewRecorder()
	KeyValidOrIP(newTestHandler(), newTestHandlerWithCode(444)).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
}

func TestKeyOrIP_IP(t *testing.T) {
	initKeyValidatorTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = ""
	resp := httptest.NewRecorder()
	KeyValidOrIP(newTestHandler(), newTestHandlerWithCode(444)).ServeHTTP(resp, req)
	assert.Equal(t, 444, resp.Code)
}
