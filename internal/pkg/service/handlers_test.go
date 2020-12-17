package service

import (
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func newTestDefH(h http.Handler) *defaultHandler {
	res := defaultHandler{h: h}
	return &res
}

func newTestQuotaH(h http.Handler, prefix, method string) *quotaHandler {
	res := quotaHandler{h: h, method: method, prefix: prefix}
	return &res
}

func TestDefaultProvider(t *testing.T) {
	h, err := NewHandler("default", newTestC(t, "default:\n  backend: http://olia.lt"), nil)
	assert.NotNil(t, h)
	assert.Nil(t, err)
	hd := h.(*defaultHandler)
	assert.NotNil(t, hd)
	assert.Equal(t, "default", hd.Name())
	assert.Equal(t, "Default handler to 'http://olia.lt'", hd.Info())
	assert.NotNil(t, hd.Handler())
	assert.Equal(t, "default", hd.Name())
	assert.True(t, hd.Valid(nil))
}

func TestDefaultProvider_Fail(t *testing.T) {
	h, err := NewHandler("default", newTestC(t, "default:\n  backend: http://"), nil)
	assert.Nil(t, h)
	assert.NotNil(t, err)
}

func TestNewHandler_Fail(t *testing.T) {
	h, err := NewHandler("default1", newTestC(t, "default1:\n  type: olia"), nil)
	assert.Nil(t, h)
	assert.NotNil(t, err)
	h, err = NewHandler("", newTestC(t, "default:\n  backend: http://olia.lt"), nil)
	assert.Nil(t, h)
	assert.NotNil(t, err)
}

// func TestMainHandlerCreate_FailPrefixURL(t *testing.T) {
// 	data := newTestData()
// 	data.Proxy.PrefixURL = ""
// 	_, err := newMainHandler(data)
// 	assert.NotNil(t, err)
// }

// func TestMainHandlerCreate_FailMethod(t *testing.T) {
// 	data := newTestData()
// 	data.Proxy.Method = ""
// 	_, err := newMainHandler(data)
// 	assert.NotNil(t, err)
// }

// func TestMainHandlerCreate_FailQuotaType(t *testing.T) {
// 	data := newTestData()
// 	data.Proxy.QuotaType = ""
// 	_, err := newMainHandler(data)
// 	assert.NotNil(t, err)
// }

// func TestMainHandlerCreate_FailAudio(t *testing.T) {
// 	data := newTestData()
// 	data.Proxy.QuotaType = "audioDuration"
// 	data.DurationService = nil
// 	_, err := newMainHandler(data)
// 	assert.NotNil(t, err)
// }

// func TestMainHandlerCreate_FailQuotaType1(t *testing.T) {
// 	data := newTestData()
// 	data.Proxy.QuotaType = "olia"
// 	_, err := newMainHandler(data)
// 	assert.NotNil(t, err)
// }

// func TestMainHandlerCreate_FailJson(t *testing.T) {
// 	data := newTestData()
// 	data.Proxy.QuotaType = "json"
// 	data.Proxy.QuotaField = ""
// 	_, err := newMainHandler(data)
// 	assert.NotNil(t, err)
// }

// func TestMainHandlerCreate_AudioJson(t *testing.T) {
// 	data := newTestData()
// 	data.Proxy.QuotaType = "audioDuration"
// 	data.Proxy.QuotaField = ""
// 	_, err := newMainHandler(data)
// 	assert.NotNil(t, err)
// }

// func TestMainHandlerCreate_Audio(t *testing.T) {
// 	data := newTestData()
// 	data.Proxy.QuotaType = "audioDuration"
// 	_, err := newMainHandler(data)
// 	assert.Nil(t, err)
// }

// func TestMainHandlerCreate_Json(t *testing.T) {
// 	data := newTestData()
// 	data.Proxy.QuotaType = "json"
// 	_, err := newMainHandler(data)
// 	assert.Nil(t, err)
// }

func newTestC(t *testing.T, configStr string) *viper.Viper {
	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(strings.NewReader(configStr))
	assert.Nil(t, err)
	return v
}
