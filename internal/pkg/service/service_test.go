package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
)

var (
	keyValidatorMock   *mocks.MockKeyValidator
	quotaValidatorMock *mocks.MockQuotaValidator
	audioLenGetterMock *mocks.MockAudioLenGetter
	dbSaverMock        *mocks.MockDBSaver
	ipManagerMock      *mocks.MockIPManager
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	keyValidatorMock = mocks.NewMockKeyValidator()
	quotaValidatorMock = mocks.NewMockQuotaValidator()
	audioLenGetterMock = mocks.NewMockAudioLenGetter()
	dbSaverMock = mocks.NewMockDBSaver()
	ipManagerMock = mocks.NewMockIPManager()
}

type testHandler struct {
	f func(http.ResponseWriter, *http.Request)
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.f(w, r)
}

func codeFunc(code int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	}
}

func TestMainHandler_Default(t *testing.T) {
	initTest(t)
	mh := mainHandler{}
	mh.data = newTestData()
	mh.def = &testHandler{f: codeFunc(222)}
	testCode(t, &mh, httptest.NewRequest("GET", "/invalid", nil), 222)
	testCode(t, &mh, httptest.NewRequest("GET", "/invalid/olia", nil), 222)
	testCode(t, &mh, httptest.NewRequest("POST", "/invalid/olia", nil), 222)
	testCode(t, &mh, httptest.NewRequest("PATCH", "/invalid/olia", nil), 222)
	testCode(t, &mh, httptest.NewRequest("DELETE", "/invalid/olia", nil), 222)
}

func TestMainHandler_Prefix(t *testing.T) {
	initTest(t)
	mh := mainHandler{}
	mh.data = newTestData()
	mh.def = &testHandler{f: codeFunc(222)}
	mh.handlers = []*hWrap{&hWrap{method: "POST", prefix: "/pref", h: &testHandler{f: codeFunc(333)}}}

	testCode(t, &mh, httptest.NewRequest("POST", "/pref", nil), 333)
	testCode(t, &mh, httptest.NewRequest("POST", "/Pref", nil), 333)
	testCode(t, &mh, httptest.NewRequest("GET", "/pref", nil), 222)
	testCode(t, &mh, httptest.NewRequest("POST", "/invalid/olia", nil), 222)
	testCode(t, &mh, httptest.NewRequest("PATCH", "/invalid/olia", nil), 222)
	testCode(t, &mh, httptest.NewRequest("DELETE", "/invalid/olia", nil), 222)
}

func TestMainHandlerCreate_FailBackend(t *testing.T) {
	data := newTestData()
	data.Proxy.BackendURL = ""
	_, err := newMainHandler(data)
	assert.NotNil(t, err)
}

func TestMainHandlerCreate_FailPrefixURL(t *testing.T) {
	data := newTestData()
	data.Proxy.PrefixURL = ""
	_, err := newMainHandler(data)
	assert.NotNil(t, err)
}

func TestMainHandlerCreate_FailMethod(t *testing.T) {
	data := newTestData()
	data.Proxy.Method = ""
	_, err := newMainHandler(data)
	assert.NotNil(t, err)
}

func TestMainHandlerCreate_FailQuotaType(t *testing.T) {
	data := newTestData()
	data.Proxy.QuotaType = ""
	_, err := newMainHandler(data)
	assert.NotNil(t, err)
}

func TestMainHandlerCreate_FailAudio(t *testing.T) {
	data := newTestData()
	data.Proxy.QuotaType = "audioDuration"
	data.DurationService = nil
	_, err := newMainHandler(data)
	assert.NotNil(t, err)
}

func TestMainHandlerCreate_FailQuotaType1(t *testing.T) {
	data := newTestData()
	data.Proxy.QuotaType = "olia"
	_, err := newMainHandler(data)
	assert.NotNil(t, err)
}

func TestMainHandlerCreate_Audio(t *testing.T) {
	data := newTestData()
	data.Proxy.QuotaType = "audioDuration"
	_, err := newMainHandler(data)
	assert.Nil(t, err)
}

func TestMainHandlerCreate_Json(t *testing.T) {
	data := newTestData()
	data.Proxy.QuotaType = "json"
	_, err := newMainHandler(data)
	assert.Nil(t, err)
}

func newTestData() *Data {
	res := &Data{DurationService: audioLenGetterMock,
		IPSaver:        ipManagerMock,
		KeyValidator:   keyValidatorMock,
		LogSaver:       dbSaverMock,
		QuotaValidator: quotaValidatorMock,
	}
	res.Proxy.BackendURL = "http://be"
	res.Proxy.Method = "POST"
	res.Proxy.PrefixURL = "/olia"
	res.Proxy.QuotaType = "json"
	res.Proxy.QuotaField = "text"
	res.Proxy.DefaultLimit = 10
	return res
}

func testCode(t *testing.T, h http.Handler, req *http.Request, code int) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)
	assert.Equal(t, code, resp.Code)
	return resp
}
