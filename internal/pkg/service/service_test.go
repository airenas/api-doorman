package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
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
	mh.data.Handlers = []HandlerWrap{newTestDefH(&testHandler{f: codeFunc(222)})}
	testCode(t, &mh, httptest.NewRequest("GET", "/invalid", nil), 222)
	testCode(t, &mh, httptest.NewRequest("GET", "/invalid/olia", nil), 222)
	testCode(t, &mh, httptest.NewRequest("POST", "/invalid/olia", nil), 222)
	testCode(t, &mh, httptest.NewRequest("PATCH", "/invalid/olia", nil), 222)
	testCode(t, &mh, httptest.NewRequest("DELETE", "/invalid/olia", nil), 222)
}

func TestMainHandler_NoDefault(t *testing.T) {
	initTest(t)
	mh := mainHandler{}
	mh.data = newTestData()
	testCode(t, &mh, httptest.NewRequest("GET", "/invalid", nil), 404)
	testCode(t, &mh, httptest.NewRequest("GET", "/invalid/olia", nil), 404)
	testCode(t, &mh, httptest.NewRequest("POST", "/invalid/olia", nil), 404)
	testCode(t, &mh, httptest.NewRequest("PATCH", "/invalid/olia", nil), 404)
	testCode(t, &mh, httptest.NewRequest("DELETE", "/invalid/olia", nil), 404)
}

func TestMainHandler_Prefix(t *testing.T) {
	initTest(t)
	mh := mainHandler{}
	mh.data = newTestData()
	mh.data.Handlers = []HandlerWrap{newTestQuotaH(&testHandler{f: codeFunc(222)}, "/pref", "GET")}

	testCode(t, &mh, httptest.NewRequest("POST", "/pref", nil), 404)
	testCode(t, &mh, httptest.NewRequest("POST", "/Pref", nil), 404)
	testCode(t, &mh, httptest.NewRequest("GET", "/pref", nil), 222)
	testCode(t, &mh, httptest.NewRequest("POST", "/invalid/olia", nil), 404)
	testCode(t, &mh, httptest.NewRequest("PATCH", "/invalid/olia", nil), 404)
	testCode(t, &mh, httptest.NewRequest("DELETE", "/invalid/olia", nil), 404)
}

func TestMainHandlerCreate(t *testing.T) {
	data := newTestData()
	data.Handlers = []HandlerWrap{newTestQuotaH(&testHandler{f: codeFunc(222)}, "/pref", "GET")}
	mh, err := newMainHandler(data)
	assert.Nil(t, err)
	assert.NotNil(t, mh)
}

func TestMainHandlerCreate_Fail(t *testing.T) {
	data := newTestData()
	mh, err := newMainHandler(data)
	assert.Nil(t, mh)
	assert.NotNil(t, err)
}

func TestGetInfo(t *testing.T) {
	th := newTestQuotaH(&testHandler{f: codeFunc(222)}, "/pref", "GET")
	th.name = "than"
	th.proxyURL = "proxy"

	hnds := []HandlerWrap{th, th}
	assert.Equal(t, "than handler GET to 'proxy', prefix: /pref\nthan handler GET to 'proxy', prefix: /pref\n", getInfo(hnds))
}

func newTestData() *Data {
	res := &Data{}
	return res
}

func testCode(t *testing.T, h http.Handler, req *http.Request, code int) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)
	assert.Equal(t, code, resp.Code)
	return resp
}
