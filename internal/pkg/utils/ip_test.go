package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/olia", nil)
	req.RemoteAddr = "192.1.1.1:8000"
	assert.Equal(t, "192.1.1.1", ExtractIP(req))
}

func TestIP_Header(t *testing.T) {
	req := httptest.NewRequest("GET", "/olia", nil)
	req.RemoteAddr = "192.1.1.1:8000"
	req.Header.Add("X-FORWARDED-FOR", "92.1.1.1:8000")
	assert.Equal(t, "92.1.1.1", ExtractIP(req))
}

func TestIP_Header_Lower(t *testing.T) {
	req := httptest.NewRequest("GET", "/olia", nil)
	req.RemoteAddr = "192.1.1.1:8000"
	req.Header.Add("x-forwarded-for", "92.1.1.1:8000")
	assert.Equal(t, "92.1.1.1", ExtractIP(req))
}

func TestIP_NoPort(t *testing.T) {
	req := httptest.NewRequest("GET", "/olia", nil)
	req.RemoteAddr = "192.1.1.1"
	assert.Equal(t, "192.1.1.1", ExtractIP(req))
}
