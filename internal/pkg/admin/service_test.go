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

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks/matchers"
	"github.com/labstack/echo/v4"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	keyCreatorMock      *mocks.MockKeyCreator
	keyRetrieverMock    *mocks.MockKeyRetriever
	oneKeyRetrieverMock *mocks.MockOneKeyRetriever
	logRetrieverMock    *mocks.MockLogRetriever
	keyUpdaterMock      *mocks.MockKeyUpdater
	prValidarorMock     *mocks.MockPrValidator
	uRestorer           *mocks.MockUsageRestorer

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
	uRestorer = mocks.NewMockUsageRestorer()
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
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"kkk"`)
	cVal := prValidarorMock.VerifyWasCalled(pegomock.Once()).Check(pegomock.AnyString()).GetCapturedArguments()
	assert.Equal(t, "pr", cVal)
}

func testToTimePtr(in time.Time) *time.Time {
	return &in
}

func TestAddKey_FailDuplicate(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.AnyString(), matchers.AnyPtrToApiKey())).ThenReturn(nil,
		mongo.WriteException{WriteErrors: []mongo.WriteError{{Code: 11000}}})
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestAddKey_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.AnyString(), matchers.AnyPtrToApiKey())).ThenReturn(nil, errors.New("err"))
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 500)
}

func TestAddKey_FailWrongField(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.AnyString(), matchers.AnyPtrToApiKey())).
		ThenReturn(nil, errors.Wrap(adminapi.ErrWrongField, "olia"))
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestAddKey_FailLimit(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.AnyString(), matchers.AnyPtrToApiKey())).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 0, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestAddKey_FailValidTo(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.AnyString(), matchers.AnyPtrToApiKey())).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(-time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
	req = httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestAddKey_FailProject(t *testing.T) {
	initTest(t)
	pegomock.When(prValidarorMock.Check(pegomock.AnyString())).ThenReturn(false)
	pegomock.When(keyCreatorMock.Create(pegomock.AnyString(), matchers.AnyPtrToApiKey())).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestUpdateKey(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.AnyString(), pegomock.AnyString(), matchers.AnyMapOfStringToInterface())).
		ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
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
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 500)
}

func TestUpdateKey_FailWrongKey(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.AnyString(), pegomock.AnyString(), matchers.AnyMapOfStringToInterface())).
		ThenReturn(nil, errors.Wrap(adminapi.ErrNoRecord, "olia"))
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestUpdateKey_FailWrongField(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.AnyString(), pegomock.AnyString(), matchers.AnyMapOfStringToInterface())).
		ThenReturn(nil, errors.Wrap(adminapi.ErrWrongField, "olia"))
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestUpdateKey_FailProject(t *testing.T) {
	initTest(t)
	pegomock.When(prValidarorMock.Check(pegomock.AnyString())).ThenReturn(false)
	pegomock.When(keyUpdaterMock.Update(pegomock.AnyString(), pegomock.AnyString(), matchers.AnyMapOfStringToInterface())).
		ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestRestore(t *testing.T) {
	initTest(t)
	pegomock.When(uRestorer.RestoreUsage(pegomock.AnyString(), pegomock.AnyBool(), pegomock.AnyString(), pegomock.AnyString())).
		ThenReturn(nil)
	req := httptest.NewRequest(http.MethodPost, "/pr/restore/m:rID", toReader(restoreReq{Error: "err msg"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	_ = testCode(t, req, http.StatusOK)
	cPr, cManual, cReq, cErr := uRestorer.VerifyWasCalled(pegomock.Once()).
		RestoreUsage(pegomock.AnyString(), pegomock.AnyBool(), pegomock.AnyString(), pegomock.AnyString()).
		GetCapturedArguments()
	assert.Equal(t, "pr", cPr)
	assert.Equal(t, true, cManual)
	assert.Equal(t, "rID", cReq)
	assert.Equal(t, "err msg", cErr)
}

func TestRestore_Fail(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodPost, "/pr/restore/rID", toReader(restoreReq{Error: "err msg"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, http.StatusBadRequest)

	req = httptest.NewRequest(http.MethodPost, "/pr/restore/m:rID", toReader(restoreReq{Error: "err msg"}))
	testCode(t, req, http.StatusBadRequest)

	req = httptest.NewRequest(http.MethodPost, "/pr/restore/m:rID", strings.NewReader(`{"error": "aaa}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, http.StatusBadRequest)

	req = httptest.NewRequest(http.MethodPost, "/pr/restore/m:rID", toReader(restoreReq{Error: "err msg"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	pegomock.When(uRestorer.RestoreUsage(pegomock.AnyString(), pegomock.AnyBool(), pegomock.AnyString(), pegomock.AnyString())).
		ThenReturn(adminapi.ErrNoRecord)
	testCode(t, req, http.StatusBadRequest)

	req = httptest.NewRequest(http.MethodPost, "/pr/restore/m:rID", toReader(restoreReq{Error: "err msg"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	pegomock.When(uRestorer.RestoreUsage(pegomock.AnyString(), pegomock.AnyBool(), pegomock.AnyString(), pegomock.AnyString())).
		ThenReturn(adminapi.ErrLogRestored)
	testCode(t, req, http.StatusConflict)

	req = httptest.NewRequest(http.MethodPost, "/pr/restore/m:rID", toReader(restoreReq{Error: "err msg"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	pegomock.When(uRestorer.RestoreUsage(pegomock.AnyString(), pegomock.AnyBool(), pegomock.AnyString(), pegomock.AnyString())).
		ThenReturn(errors.New("olia"))
	testCode(t, req, http.StatusInternalServerError)
}

func toReader(key interface{}) io.Reader {
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
		UsageRestorer:    uRestorer,
	}
	return res
}

func testCode(t *testing.T, req *http.Request, code int) *httptest.ResponseRecorder {
	t.Helper()
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

func Test_parseRequestID(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    string
		wantM   bool
		wantErr bool
	}{
		{name: "Manual", args: "m:id", want: "id", wantM: true, wantErr: false},
		{name: "Manual false", args: ":id", want: "id", wantM: false, wantErr: false},
		{name: "Fails", args: "id", wantErr: true},
		{name: "Fails2", args: "::id", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parseRequestID(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRequestID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseRequestID() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.wantM {
				t.Errorf("parseRequestID() got1 = %v, want %v", got1, tt.wantM)
			}
		})
	}
}
