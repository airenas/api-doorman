package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripPrefix(t *testing.T) {
	req := httptest.NewRequest("POST", "/duration/olia", nil)
	resp := httptest.NewRecorder()
	StripPrefix(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/olia", req.URL.String())
	}), "/duration").ServeHTTP(resp, req)
}

func TestStripPrefix_NoMatch(t *testing.T) {
	req := httptest.NewRequest("POST", "/duration/olia", nil)
	resp := httptest.NewRecorder()
	StripPrefix(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/duration/olia", req.URL.String())
	}), "/olia").ServeHTTP(resp, req)
}
