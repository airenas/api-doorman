package service

import (
	"net/http"
)

func newTestDefH(h http.Handler) *defaultHandler {
	res := defaultHandler{h: h}
	return &res
}

func newTestQuotaH(h http.Handler, prefix, method string) *quotaHandler {
	res := quotaHandler{h: h, method: method, prefix: prefix}
	return &res
}

// func TestMainHandlerCreate_FailBackend(t *testing.T) {
// 	data := newTestData()
// 	data.Proxy.BackendURL = ""
// 	_, err := newMainHandler(data)
// 	assert.NotNil(t, err)
// 	data.Proxy.BackendURL = "http://"
// 	_, err = newMainHandler(data)
// 	assert.NotNil(t, err)
// }

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
