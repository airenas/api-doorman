package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFillOutHeader_None(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Tags = []string{"olia:15"}
	resp := httptest.NewRecorder()
	FillOutHeader(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, 0, len(resp.Header()))
}

func TestFillOutHeader(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Tags = []string{tagStartValue + "olia:15"}
	resp := httptest.NewRecorder()
	FillOutHeader(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "15", resp.Header().Get("olia"))
}

func TestFillOutHeader_Several(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Tags = []string{tagStartValue + "olia:15", tagStartValue + "xkkk: 12:16"}
	resp := httptest.NewRecorder()
	FillOutHeader(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, "15", resp.Header().Get("olia"))
	assert.Equal(t, "12:16", resp.Header().Get("xkkk"))
}

func TestFillOutHeader_Fail(t *testing.T) {
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Tags = []string{tagStartValue + "olia=15"}
	resp := httptest.NewRecorder()
	FillOutHeader(newTestHandler()).ServeHTTP(resp, req)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}
