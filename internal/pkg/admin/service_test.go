package admin

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
)

var (
	keyCreatorMock      *mocks.MockKeyCreator
	keyRetrieverMock    *mocks.MockKeyRetriever
	oneKeyRetrieverMock *mocks.MockOneKeyRetriever
	logRetrieverMock    *mocks.MockLogRetriever
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	keyCreatorMock = mocks.NewMockKeyCreator()
	keyRetrieverMock = mocks.NewMockKeyRetriever()
	oneKeyRetrieverMock = mocks.NewMockOneKeyRetriever()
	logRetrieverMock = mocks.NewMockLogRetriever()
	//	pegomock.When(recognizerMapMock.Get(pegomock.AnyString())).ThenReturn("recID", nil)
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

func newTestRouter() *mux.Router {
	return NewRouter(newTestData())
}

func newTestData() *Data {
	res := &Data{KeySaver: keyCreatorMock,
		KeyGetter:    keyRetrieverMock,
		OneKeyGetter: oneKeyRetrieverMock,
		LogGetter:    logRetrieverMock,
	}
	return res
}

func testCode(t *testing.T, req *http.Request, code int) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	newTestRouter().ServeHTTP(resp, req)
	assert.Equal(t, code, resp.Code)
	return resp
}
