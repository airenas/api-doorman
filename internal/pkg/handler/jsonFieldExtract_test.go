package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", strings.NewReader(`{"body":"olia"}`)))
	resp := httptest.NewRecorder()

	TakeJSON(newTestHandler(), "body").ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "olia", ctx.Value)
}

func TestJSON_Empty(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", strings.NewReader(`{"body":"olia"}`)))
	resp := httptest.NewRecorder()

	TakeJSON(newTestHandler(), "body1").ServeHTTP(resp, req)
	assert.Equal(t, "", ctx.Value)
	assert.Equal(t, testCode, resp.Code)
}

func TestJSON_Fail(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", strings.NewReader(`{"body":"olia}`)))
	resp := httptest.NewRecorder()

	TakeJSON(newTestHandler(), "body1").ServeHTTP(resp, req)
	assert.Equal(t, "", ctx.Value)
	assert.Equal(t, 400, resp.Code)
}
