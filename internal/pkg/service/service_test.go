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

func newTestData() *Data {
	res := &Data{DurationService: audioLenGetterMock,
		IPSaver:        ipManagerMock,
		KeyValidator:   keyValidatorMock,
		LogSaver:       dbSaverMock,
		QuotaValidator: quotaValidatorMock,
	}
	return res
}

func testCode(t *testing.T, h http.Handler, req *http.Request, code int) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)
	assert.Equal(t, code, resp.Code)
	return resp
}
