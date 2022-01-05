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

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/labstack/echo/v4"
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
	prValidarorMock     *mocks.MockPrValidator

	tData *Data
	tEcho *echo.Echo
	tResp *httptest.ResponseRecorder
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	keyCreatorMock = mocks.NewMockKeyCreator()
	keyRetrieverMock = mocks.NewMockKeyRetriever()
	oneKeyRetrieverMock = mocks.NewMockOneKeyRetriever()
	logRetrieverMock = mocks.NewMockLogRetriever()
	keyUpdaterMock = mocks.NewMockKeyUpdater()
	prValidarorMock = mocks.NewMockPrValidator()
	pegomock.When(prValidarorMock.Check(pegomock.AnyString())).ThenReturn(true)

	tData = newTestData()
	tEcho = initRoutes(tData)
	tResp = httptest.NewRecorder()
}

func TestWrongPath(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("GET", "/invalid", nil)
	testCode(t, req, 404)
}

func TestKeyList(t *testing.T) {
	initTest(t)
	pegomock.When(keyRetrieverMock.List(pegomock.AnyString())).ThenReturn([]*adminapi.Key{}, nil)
	req := httptest.NewRequest("GET", "/pr/key-list", nil)
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "[]\n", string(bytes))
	cVal := prValidarorMock.VerifyWasCalled(pegomock.Once()).Check(pegomock.AnyString()).GetCapturedArguments()
	assert.Equal(t, "pr", cVal)
}

func TestKeyList_Returns(t *testing.T) {
	initTest(t)
	pegomock.When(keyRetrieverMock.List(pegomock.AnyString())).ThenReturn([]*adminapi.Key{{Key: "olia"}}, nil)
	req := httptest.NewRequest("GET", "/pr/key-list", nil)
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"olia"`)
}

func TestKeyList_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(keyRetrieverMock.List(pegomock.AnyString())).ThenReturn(nil, errors.New("olia"))
	req := httptest.NewRequest("GET", "/pr/key-list", nil)
	testCode(t, req, 500)
}

func TestKeyList_FailProject(t *testing.T) {
	initTest(t)
	pegomock.When(prValidarorMock.Check(pegomock.AnyString())).ThenReturn(false)
	req := httptest.NewRequest("GET", "/pr/key-list", nil)
	testCode(t, req, 400)
}

func TestKey(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.AnyString(), pegomock.EqString("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("GET", "/pr/key/kkk", nil)
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"kkk"`)
	cVal := prValidarorMock.VerifyWasCalled(pegomock.Once()).Check(pegomock.AnyString()).GetCapturedArguments()
	assert.Equal(t, "pr", cVal)
}

func TestKey_ReturnsFull(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.AnyString(), pegomock.EqString("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	pegomock.When(logRetrieverMock.Get(pegomock.AnyString(), pegomock.EqString("kkk"))).ThenReturn([]*adminapi.Log{{IP: "101010"}}, nil)
	req := httptest.NewRequest("GET", "/pr/key/kkk?full=1", nil)
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"ip":"101010"`)
}

func TestKey_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.AnyString(), pegomock.EqString("kkk"))).ThenReturn(nil, errors.New("fail"))
	req := httptest.NewRequest("GET", "/pr/key/kkk", nil)
	testCode(t, req, 500)
}

func TestKey_FailFull(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.AnyString(), pegomock.EqString("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	pegomock.When(logRetrieverMock.Get(pegomock.AnyString(), pegomock.EqString("kkk"))).ThenReturn(nil, errors.New("fail"))
	req := httptest.NewRequest("GET", "/pr/key/kkk?full=1", nil)
	testCode(t, req, 500)
}

func TestKey_FailNoKey(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.AnyString(), pegomock.EqString("kkk"))).ThenReturn(nil, adminapi.ErrNoRecord)
	req := httptest.NewRequest("GET", "/pr/key/kkk", nil)
	testCode(t, req, 400)
}

func TestKey_FailProject(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.AnyString(), pegomock.EqString("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	pegomock.When(prValidarorMock.Check(pegomock.AnyString())).ThenReturn(false)
	req := httptest.NewRequest("GET", "/pr/key-list", nil)
	testCode(t, req, 400)
}

func TestAddKey(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.AnyString(), matchers.AnyPtrToApiKey())).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(time.Minute)}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"kkk"`)
	cVal := prValidarorMock.VerifyWasCalled(pegomock.Once()).Check(pegomock.AnyString()).GetCapturedArguments()
	assert.Equal(t, "pr", cVal)
}

func TestAddKey_FailDuplicate(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.AnyString(), matchers.AnyPtrToApiKey())).ThenReturn(nil,
		mongo.WriteException{WriteErrors: []mongo.WriteError{{Code: 11000}}})
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(time.Minute)}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestAddKey_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.AnyString(), matchers.AnyPtrToApiKey())).ThenReturn(nil, errors.New("err"))
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(time.Minute)}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 500)
}

func TestAddKey_FailWrongField(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.AnyString(), matchers.AnyPtrToApiKey())).
		ThenReturn(nil, errors.Wrap(adminapi.ErrWrongField, "olia"))
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(time.Minute)}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestAddKey_FailLimit(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.AnyString(), matchers.AnyPtrToApiKey())).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 0, ValidTo: time.Now().Add(time.Minute)}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestAddKey_FailValidTo(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.AnyString(), matchers.AnyPtrToApiKey())).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(-time.Minute)}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestAddKey_FailProject(t *testing.T) {
	initTest(t)
	pegomock.When(prValidarorMock.Check(pegomock.AnyString())).ThenReturn(false)
	pegomock.When(keyCreatorMock.Create(pegomock.AnyString(), matchers.AnyPtrToApiKey())).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(time.Minute)}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestUpdateKey(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.AnyString(), pegomock.AnyString(), matchers.AnyMapOfStringToInterface())).
		ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(time.Minute)}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"kkk"`)
	cPr, cKey, cMap := keyUpdaterMock.VerifyWasCalled(pegomock.Once()).Update(pegomock.AnyString(), pegomock.AnyString(), matchers.AnyMapOfStringToInterface()).
		GetCapturedArguments()
	assert.Equal(t, "pr", cPr)
	assert.Equal(t, "kkk", cKey)
	_, ok := cMap["limit"]
	assert.True(t, ok)
	_, ok = cMap["validTo"]
	assert.True(t, ok)
	cVal := prValidarorMock.VerifyWasCalled(pegomock.Once()).Check(pegomock.AnyString()).GetCapturedArguments()
	assert.Equal(t, "pr", cVal)
}

func TestUpdateKey_FailInput(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.AnyString(), pegomock.AnyString(), matchers.AnyMapOfStringToInterface())).
		ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", strings.NewReader("{{{"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestUpdateKey_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.AnyString(), pegomock.AnyString(), matchers.AnyMapOfStringToInterface())).
		ThenReturn(nil, errors.New("olia"))
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(time.Minute)}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 500)
}

func TestUpdateKey_FailWrongKey(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.AnyString(), pegomock.AnyString(), matchers.AnyMapOfStringToInterface())).
		ThenReturn(nil, errors.Wrap(adminapi.ErrNoRecord, "olia"))
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(time.Minute)}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestUpdateKey_FailWrongField(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.AnyString(), pegomock.AnyString(), matchers.AnyMapOfStringToInterface())).
		ThenReturn(nil, errors.Wrap(adminapi.ErrWrongField, "olia"))
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(time.Minute)}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestUpdateKey_FailProject(t *testing.T) {
	initTest(t)
	pegomock.When(prValidarorMock.Check(pegomock.AnyString())).ThenReturn(false)
	pegomock.When(keyUpdaterMock.Update(pegomock.AnyString(), pegomock.AnyString(), matchers.AnyMapOfStringToInterface())).
		ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: time.Now().Add(time.Minute)}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func toReader(key adminapi.Key) io.Reader {
	bytes, _ := json.Marshal(key)
	return strings.NewReader(string(bytes))
}

func newTestData() *Data {
	res := &Data{KeySaver: keyCreatorMock,
		KeyGetter:        keyRetrieverMock,
		OneKeyGetter:     oneKeyRetrieverMock,
		LogGetter:        logRetrieverMock,
		OneKeyUpdater:    keyUpdaterMock,
		ProjectValidator: prValidarorMock,
	}
	return res
}

func testCode(t *testing.T, req *http.Request, code int) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	tEcho.ServeHTTP(resp, req)
	assert.Equal(t, code, resp.Code)
	return resp
}

func TestLive(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodGet, "/live", nil)

	tEcho.ServeHTTP(tResp, req)
	assert.Equal(t, http.StatusOK, tResp.Code)
	assert.Equal(t, `{"service":"OK"}`, tResp.Body.String())
}
