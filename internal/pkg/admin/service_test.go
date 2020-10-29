package admin

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks/matchers"
)

var (
	keyCreatorMock      *mocks.MockKeyCreator
	keyRetrieverMock    *mocks.MockKeyRetriever
	oneKeyRetrieverMock *mocks.MockOneKeyRetriever
	logRetrieverMock    *mocks.MockLogRetriever
	keyUpdaterMock      *mocks.MockKeyUpdater
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	keyCreatorMock = mocks.NewMockKeyCreator()
	keyRetrieverMock = mocks.NewMockKeyRetriever()
	oneKeyRetrieverMock = mocks.NewMockOneKeyRetriever()
	logRetrieverMock = mocks.NewMockLogRetriever()
	keyUpdaterMock = mocks.NewMockKeyUpdater()
}

func TestWrongPath(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("GET", "/invalid", nil)
	testCode(t, req, 404)
}

func TestKeyList(t *testing.T) {
	initTest(t)
	pegomock.When(keyRetrieverMock.List()).ThenReturn([]*adminapi.Key{}, nil)
	req := httptest.NewRequest("GET", "/key-list", nil)
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "[]\n", string(bytes))
}

func TestKeyList_Returns(t *testing.T) {
	initTest(t)
	pegomock.When(keyRetrieverMock.List()).ThenReturn([]*adminapi.Key{&adminapi.Key{Key: "olia"}}, nil)
	req := httptest.NewRequest("GET", "/key-list", nil)
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"olia"`)
}

func TestKeyList_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(keyRetrieverMock.List()).ThenReturn(nil, errors.New("olia"))
	req := httptest.NewRequest("GET", "/key-list", nil)
	testCode(t, req, 500)
}

func TestKey(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.EqString("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("GET", "/key/kkk", nil)
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"kkk"`)
}

func TestKey_ReturnsFull(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.EqString("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	pegomock.When(logRetrieverMock.Get(pegomock.EqString("kkk"))).ThenReturn([]*adminapi.Log{&adminapi.Log{IP: "101010"}}, nil)
	req := httptest.NewRequest("GET", "/key/kkk?full=1", nil)
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"ip":"101010"`)
}

func TestKey_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.EqString("kkk"))).ThenReturn(nil, errors.New("fail"))
	req := httptest.NewRequest("GET", "/key/kkk", nil)
	testCode(t, req, 500)
}

func TestKey_FailFull(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.EqString("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	pegomock.When(logRetrieverMock.Get(pegomock.EqString("kkk"))).ThenReturn(nil, errors.New("fail"))
	req := httptest.NewRequest("GET", "/key/kkk?full=1", nil)
	testCode(t, req, 500)
}

func TestKey_FailNoKey(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.EqString("kkk"))).ThenReturn(nil, nil)
	req := httptest.NewRequest("GET", "/key/kkk", nil)
	testCode(t, req, 400)
}

func TestAddKey(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(matchers.AnyPtrToApiKey())).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("POST", "/key", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(time.Minute)}))
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"kkk"`)
}

func TestAddKey_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(matchers.AnyPtrToApiKey())).ThenReturn(nil, errors.New("err"))
	req := httptest.NewRequest("POST", "/key", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(time.Minute)}))
	testCode(t, req, 500)
}

func TestAddKey_FailLimit(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(matchers.AnyPtrToApiKey())).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("POST", "/key", toReader(adminapi.Key{Limit: 0, ValidTo: time.Now().Add(time.Minute)}))
	testCode(t, req, 400)
}

func TestAddKey_FailValidTo(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(matchers.AnyPtrToApiKey())).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("POST", "/key", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(-time.Minute)}))
	testCode(t, req, 400)
}

func TestUpdateKey(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.AnyString(), matchers.AnyMapOfStringToInterface())).
		ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("PATCH", "/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(time.Minute)}))
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"kkk"`)
	cKey, cMap := keyUpdaterMock.VerifyWasCalled(pegomock.Once()).Update(pegomock.AnyString(), matchers.AnyMapOfStringToInterface()).
		GetCapturedArguments()
	assert.Equal(t, "kkk", cKey)
	_, ok := cMap["limit"]
	assert.True(t, ok)
	_, ok = cMap["validTo"]
	assert.True(t, ok)
}

func TestUpdateKey_FailInput(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.AnyString(), matchers.AnyMapOfStringToInterface())).
		ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("PATCH", "/key/kkk", strings.NewReader("{{{"))
	testCode(t, req, 400)
}

func TestUpdateKey_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.AnyString(), matchers.AnyMapOfStringToInterface())).
		ThenReturn(nil, errors.New("olia"))
	req := httptest.NewRequest("PATCH", "/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(time.Minute)}))
	testCode(t, req, 500)
}

func newTestRouter() *mux.Router {
	return NewRouter(newTestData())
}

func toReader(key adminapi.Key) io.Reader {
	bytes, _ := json.Marshal(key)
	return strings.NewReader(string(bytes))
}

func newTestData() *Data {
	res := &Data{KeySaver: keyCreatorMock,
		KeyGetter:     keyRetrieverMock,
		OneKeyGetter:  oneKeyRetrieverMock,
		LogGetter:     logRetrieverMock,
		OneKeyUpdater: keyUpdaterMock,
	}
	return res
}

func testCode(t *testing.T, req *http.Request, code int) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	newTestRouter().ServeHTTP(resp, req)
	assert.Equal(t, code, resp.Code)
	return resp
}
