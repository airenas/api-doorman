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

func TestValidateIPsCIDR(t *testing.T) {
	assert.Nil(t, ValidateIPsCIDR(""))
	assert.Nil(t, ValidateIPsCIDR("1.1.1.1/32"))
	assert.Nil(t, ValidateIPsCIDR("1.1.1.1/32,2.2.2.2/24"))
	assert.NotNil(t, ValidateIPsCIDR("aaaa"))
	assert.NotNil(t, ValidateIPsCIDR("10.10/11/11"))
	assert.NotNil(t, ValidateIPsCIDR("1.1.1.1/32,2.2.2.2/24/"))
	assert.NotNil(t, ValidateIPsCIDR("1.1.1.1"))
}

func TestValidateIPInWhiteList_Empty(t *testing.T) {
	v, err := ValidateIPInWhiteList("", "")
	assert.Nil(t, err)
	assert.True(t, v)
}

func TestValidateIPInWhiteList_Valid(t *testing.T) {
	v, err := ValidateIPInWhiteList("1.1.1.1/32", "1.1.1.1")
	assert.Nil(t, err)
	assert.True(t, v)
	v, _ = ValidateIPInWhiteList("1.1.1.1/32", "1.1.1.2")
	assert.False(t, v)
	v, _ = ValidateIPInWhiteList("1.1.1.1/24", "1.1.1.2")
	assert.True(t, v)
	v, _ = ValidateIPInWhiteList("1.1.1.1/32,1.1.1.1/16", "1.1.2.2")
	assert.True(t, v)
	v, _ = ValidateIPInWhiteList("1.1.1.1/32,1.1.1.1/16,255.1.1.1/8", "255.254.55.222")
	assert.True(t, v)
}

func TestValidateIPInWhiteList_Errors(t *testing.T) {
	v, err := ValidateIPInWhiteList("1.1.1.1/32", "1.1.1")
	assert.NotNil(t, err)
	assert.False(t, v)
	_, err = ValidateIPInWhiteList("1.1.1.1/32", "")
	assert.NotNil(t, err)
	_, err = ValidateIPInWhiteList("1.1.1./32", "1.1.1.1")
	assert.NotNil(t, err)
	_, err = ValidateIPInWhiteList("1.1.1.1", "1.1.1.1")
	assert.NotNil(t, err)
	_, err = ValidateIPInWhiteList("1.1.1.1/32,11.1.1.1//", "1.1.1.2")
	assert.NotNil(t, err)
}
