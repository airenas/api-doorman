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

func TestNewExtractor(t *testing.T) {
	e, err := NewIPExtractor("lastForwardFor")
	assert.NotNil(t, e)
	assert.Nil(t, err)
	e, err = NewIPExtractor("firstForwardFor")
	assert.NotNil(t, e)
	assert.Nil(t, err)
}

func TestNewExtractor_Fail(t *testing.T) {
	e, err := NewIPExtractor("")
	assert.Nil(t, e)
	assert.NotNil(t, err)

	e, err = NewIPExtractor("olia")
	assert.Nil(t, e)
	assert.NotNil(t, err)
}

func TestCreatesDefault(t *testing.T) {
	req := httptest.NewRequest("GET", "/olia", nil)
	req.Header.Add("X-FORWARDED-FOR", "92.1.1.1:8000,92.1.1.2:8000, 92.1.1.3:8000")
	DefaultIPExtractor = nil
	assert.Equal(t, "92.1.1.1", ExtractIP(req))
	DefaultIPExtractor = &lastForwardFor{}
	assert.Equal(t, "92.1.1.3", ExtractIP(req))
	DefaultIPExtractor = &firstForwardFor{}
	assert.Equal(t, "92.1.1.1", ExtractIP(req))
}

func TestWithSaces(t *testing.T) {
	req := httptest.NewRequest("GET", "/olia", nil)
	req.Header.Add("X-FORWARDED-FOR", "       92.1.1.1:8000    ,92.1.1.2:8000      , 92.1.1.3:8000")
	assert.Equal(t, "92.1.1.1", ExtractIP(req))
}

func TestNoPort(t *testing.T) {
	req := httptest.NewRequest("GET", "/olia", nil)
	req.Header.Add("X-FORWARDED-FOR", "92.1.1.1, 92.1.1.3:8000")
	assert.Equal(t, "92.1.1.1", ExtractIP(req))
}

func TestGetIPHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/olia", nil)
	req.RemoteAddr = "192.1.1.1:8000"
	req.Header.Add("X-FORWARDED-FOR", "92.1.1.1:8000")
	assert.Equal(t, "92.1.1.1:8000", GetIPHeader(req))
}
