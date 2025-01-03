package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/admin/api"
	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/labstack/echo/v4"
	"github.com/petergtz/pegomock/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	keyCreatorMock      *mocks.MockKeyCreator
	keyRetrieverMock    *mocks.MockKeyRetriever
	oneKeyRetrieverMock *mocks.MockOneKeyRetriever
	logRetrieverMock    *mocks.MockLogProvider
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
	logRetrieverMock = mocks.NewMockLogProvider()
	keyUpdaterMock = mocks.NewMockKeyUpdater()
	prValidarorMock = mocks.NewMockPrValidator()
	uRestorer = mocks.NewMockUsageRestorer()
	pegomock.When(prValidarorMock.Check(pegomock.Any[string]())).ThenReturn(true)

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
	pegomock.When(keyRetrieverMock.List(pegomock.Any[string]())).ThenReturn([]*adminapi.Key{}, nil)
	req := httptest.NewRequest("GET", "/pr/key-list", nil)
	resp := testCode(t, req, 200)
	bytes, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "[]\n", string(bytes))
	cVal := prValidarorMock.VerifyWasCalled(pegomock.Once()).Check(pegomock.Any[string]()).GetCapturedArguments()
	assert.Equal(t, "pr", cVal)
}

func TestKeyList_Returns(t *testing.T) {
	initTest(t)
	pegomock.When(keyRetrieverMock.List(pegomock.Any[string]())).ThenReturn([]*adminapi.Key{{Key: "olia"}}, nil)
	req := httptest.NewRequest("GET", "/pr/key-list", nil)
	resp := testCode(t, req, 200)
	bytes, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"olia"`)
}

func TestKeyList_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(keyRetrieverMock.List(pegomock.Any[string]())).ThenReturn(nil, errors.New("olia"))
	req := httptest.NewRequest("GET", "/pr/key-list", nil)
	testCode(t, req, 500)
}

func TestKeyList_FailProject(t *testing.T) {
	initTest(t)
	pegomock.When(prValidarorMock.Check(pegomock.Any[string]())).ThenReturn(false)
	req := httptest.NewRequest("GET", "/pr/key-list", nil)
	testCode(t, req, 400)
}

func TestKey(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.Any[string](), pegomock.Eq("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("GET", "/pr/key/kkk", nil)
	resp := testCode(t, req, 200)
	bytes, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"kkk"`)
	cVal := prValidarorMock.VerifyWasCalled(pegomock.Once()).Check(pegomock.Any[string]()).GetCapturedArguments()
	assert.Equal(t, "pr", cVal)
}

func TestKey_ReturnsFull(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.Any[string](), pegomock.Eq("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	pegomock.When(logRetrieverMock.Get(pegomock.Any[string](), pegomock.Eq("kkk"))).ThenReturn([]*adminapi.Log{{IP: "101010"}}, nil)
	req := httptest.NewRequest("GET", "/pr/key/kkk?full=1", nil)
	resp := testCode(t, req, 200)
	bytes, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"ip":"101010"`)
}

func TestKey_ReturnsFullNyKeyID(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.Any[string](), pegomock.Eq("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk", ID: "olia"}, nil)
	pegomock.When(logRetrieverMock.Get(pegomock.Any[string](), pegomock.Eq("kkk"))).ThenReturn([]*adminapi.Log{{IP: "101010"}}, nil)
	pegomock.When(logRetrieverMock.Get(pegomock.Any[string](), pegomock.Eq("olia"))).ThenReturn([]*adminapi.Log{{IP: "101011"}}, nil)
	req := httptest.NewRequest("GET", "/pr/key/kkk?full=1", nil)
	resp := testCode(t, req, 200)
	bytes, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"ip":"101010"`)
	assert.Contains(t, string(bytes), `"ip":"101011"`)
}

func TestKey_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.Any[string](), pegomock.Eq("kkk"))).ThenReturn(nil, errors.New("fail"))
	req := httptest.NewRequest("GET", "/pr/key/kkk", nil)
	testCode(t, req, 500)
}

func TestKey_FailIDCall(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.Any[string](), pegomock.Eq("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk", ID: "olia"}, nil)
	pegomock.When(logRetrieverMock.Get(pegomock.Any[string](), pegomock.Eq("kkk"))).ThenReturn([]*adminapi.Log{{IP: "101010"}}, nil)
	pegomock.When(logRetrieverMock.Get(pegomock.Any[string](), pegomock.Eq("olia"))).ThenReturn(nil, errors.New("fail"))
	req := httptest.NewRequest("GET", "/pr/key/kkk?full=1", nil)
	testCode(t, req, 500)
}

func TestKey_FailFull(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.Any[string](), pegomock.Eq("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	pegomock.When(logRetrieverMock.Get(pegomock.Any[string](), pegomock.Eq("kkk"))).ThenReturn(nil, errors.New("fail"))
	req := httptest.NewRequest("GET", "/pr/key/kkk?full=1", nil)
	testCode(t, req, 500)
}

func TestKey_FailNoKey(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.Any[string](), pegomock.Eq("kkk"))).ThenReturn(nil, adminapi.ErrNoRecord)
	req := httptest.NewRequest("GET", "/pr/key/kkk", nil)
	testCode(t, req, 400)
}

func TestKey_FailProject(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.Any[string](), pegomock.Eq("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	pegomock.When(prValidarorMock.Check(pegomock.Any[string]())).ThenReturn(false)
	req := httptest.NewRequest("GET", "/pr/key-list", nil)
	testCode(t, req, 400)
}

func TestAddKey(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.Any[string](), pegomock.Any[*api.Key]())).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp := testCode(t, req, 200)
	bytes, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"kkk"`)
	cVal := prValidarorMock.VerifyWasCalled(pegomock.Once()).Check(pegomock.Any[string]()).GetCapturedArguments()
	assert.Equal(t, "pr", cVal)
}

func testToTimePtr(in time.Time) *time.Time {
	return &in
}

func TestAddKey_FailDuplicate(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.Any[string](), pegomock.Any[*api.Key]())).ThenReturn(nil,
		mongo.WriteException{WriteErrors: []mongo.WriteError{{Code: 11000}}})
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestAddKey_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.Any[string](), pegomock.Any[*api.Key]())).ThenReturn(nil, errors.New("err"))
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 500)
}

func TestAddKey_FailWrongField(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.Any[string](), pegomock.Any[*api.Key]())).
		ThenReturn(nil, errors.Wrap(adminapi.ErrWrongField, "olia"))
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestAddKey_FailLimit(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.Any[string](), pegomock.Any[*api.Key]())).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 0, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestAddKey_FailValidTo(t *testing.T) {
	initTest(t)
	pegomock.When(keyCreatorMock.Create(pegomock.Any[string](), pegomock.Any[*api.Key]())).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(-time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
	req = httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestAddKey_FailProject(t *testing.T) {
	initTest(t)
	pegomock.When(prValidarorMock.Check(pegomock.Any[string]())).ThenReturn(false)
	pegomock.When(keyCreatorMock.Create(pegomock.Any[string](), pegomock.Any[*api.Key]())).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("POST", "/pr/key", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestUpdateKey(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.Any[string](), pegomock.Any[string](), pegomock.Any[map[string]interface{}]())).
		ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp := testCode(t, req, 200)
	bytes, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"kkk"`)
	cPr, cKey, cMap := keyUpdaterMock.VerifyWasCalled(pegomock.Once()).Update(pegomock.Any[string](), pegomock.Any[string](), pegomock.Any[map[string]interface{}]()).
		GetCapturedArguments()
	assert.Equal(t, "pr", cPr)
	assert.Equal(t, "kkk", cKey)
	_, ok := cMap["limit"]
	assert.True(t, ok)
	_, ok = cMap["validTo"]
	assert.True(t, ok)
	cVal := prValidarorMock.VerifyWasCalled(pegomock.Once()).Check(pegomock.Any[string]()).GetCapturedArguments()
	assert.Equal(t, "pr", cVal)
}

func TestUpdateKey_FailInput(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.Any[string](), pegomock.Any[string](), pegomock.Any[map[string]interface{}]())).
		ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", strings.NewReader("{{{"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestUpdateKey_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.Any[string](), pegomock.Any[string](), pegomock.Any[map[string]interface{}]())).
		ThenReturn(nil, errors.New("olia"))
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 500)
}

func TestUpdateKey_FailWrongKey(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.Any[string](), pegomock.Any[string](), pegomock.Any[map[string]interface{}]())).
		ThenReturn(nil, errors.Wrap(adminapi.ErrNoRecord, "olia"))
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestUpdateKey_FailWrongField(t *testing.T) {
	initTest(t)
	pegomock.When(keyUpdaterMock.Update(pegomock.Any[string](), pegomock.Any[string](), pegomock.Any[map[string]interface{}]())).
		ThenReturn(nil, errors.Wrap(adminapi.ErrWrongField, "olia"))
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestUpdateKey_FailProject(t *testing.T) {
	initTest(t)
	pegomock.When(prValidarorMock.Check(pegomock.Any[string]())).ThenReturn(false)
	pegomock.When(keyUpdaterMock.Update(pegomock.Any[string](), pegomock.Any[string](), pegomock.Any[map[string]interface{}]())).
		ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("PATCH", "/pr/key/kkk", toReader(adminapi.Key{Limit: 10, ValidTo: testToTimePtr(time.Now().Add(time.Minute))}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	testCode(t, req, 400)
}

func TestRestore(t *testing.T) {
	initTest(t)
	pegomock.When(uRestorer.RestoreUsage(pegomock.Any[string](), pegomock.Any[bool](), pegomock.Any[string](), pegomock.Any[string]())).
		ThenReturn(nil)
	req := httptest.NewRequest(http.MethodPost, "/pr/restore/m:rID", toReader(restoreReq{Error: "err msg"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	_ = testCode(t, req, http.StatusOK)
	cPr, cManual, cReq, cErr := uRestorer.VerifyWasCalled(pegomock.Once()).
		RestoreUsage(pegomock.Any[string](), pegomock.Any[bool](), pegomock.Any[string](), pegomock.Any[string]()).
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
	pegomock.When(uRestorer.RestoreUsage(pegomock.Any[string](), pegomock.Any[bool](), pegomock.Any[string](), pegomock.Any[string]())).
		ThenReturn(adminapi.ErrNoRecord)
	testCode(t, req, http.StatusBadRequest)

	req = httptest.NewRequest(http.MethodPost, "/pr/restore/m:rID", toReader(restoreReq{Error: "err msg"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	pegomock.When(uRestorer.RestoreUsage(pegomock.Any[string](), pegomock.Any[bool](), pegomock.Any[string](), pegomock.Any[string]())).
		ThenReturn(adminapi.ErrLogRestored)
	testCode(t, req, http.StatusConflict)

	req = httptest.NewRequest(http.MethodPost, "/pr/restore/m:rID", toReader(restoreReq{Error: "err msg"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	pegomock.When(uRestorer.RestoreUsage(pegomock.Any[string](), pegomock.Any[bool](), pegomock.Any[string](), pegomock.Any[string]())).
		ThenReturn(errors.New("olia"))
	testCode(t, req, http.StatusInternalServerError)
}

func TestLogList(t *testing.T) {
	initTest(t)
	pegomock.When(logRetrieverMock.List(pegomock.Any[string](), pegomock.Any[time.Time]())).
		ThenReturn([]*adminapi.Log{{KeyID: "1", ResponseCode: 200}, {KeyID: "2", ResponseCode: 400}}, nil)
	req := httptest.NewRequest(http.MethodGet, "/pr/log?to=2023-01-02T15:04:05Z", nil)
	resp := testCode(t, req, http.StatusOK)
	bytes, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"keyID":"1"`)
	assert.Contains(t, string(bytes), `"keyID":"2"`)
	assert.Contains(t, string(bytes), `"response":400`)
	cPr, cTime := logRetrieverMock.VerifyWasCalled(pegomock.Once()).
		List(pegomock.Any[string](), pegomock.Any[time.Time]()).
		GetCapturedArguments()
	assert.Equal(t, "pr", cPr)
	assert.Equal(t, time.Date(2023, time.January, 2, 15, 04, 05, 0, time.UTC), cTime)
}

func TestLogList_FailDate(t *testing.T) {
	initTest(t)
	pegomock.When(logRetrieverMock.List(pegomock.Any[string](), pegomock.Any[time.Time]())).
		ThenReturn([]*adminapi.Log{{KeyID: "1", ResponseCode: 200}, {KeyID: "2", ResponseCode: 400}}, nil)
	req := httptest.NewRequest(http.MethodGet, "/pr/log", nil)
	testCode(t, req, http.StatusBadRequest)
}

func TestLogList_FailDB(t *testing.T) {
	initTest(t)
	pegomock.When(logRetrieverMock.List(pegomock.Any[string](), pegomock.Any[time.Time]())).
		ThenReturn(nil, fmt.Errorf("err"))
	req := httptest.NewRequest(http.MethodGet, "/pr/log?to=2023-01-02T15:04:05Z", nil)
	testCode(t, req, http.StatusInternalServerError)
}

func TestLogDelete(t *testing.T) {
	initTest(t)
	pegomock.When(logRetrieverMock.Delete(pegomock.Any[string](), pegomock.Any[time.Time]())).ThenReturn(10, nil)
	req := httptest.NewRequest(http.MethodDelete, "/pr/log?to=2023-01-02T15:04:05Z", nil)
	resp := testCode(t, req, http.StatusOK)
	bytes, _ := io.ReadAll(resp.Body)
	assert.Equal(t, `{"deleted":10}`, string(bytes))
	cPr, cTime := logRetrieverMock.VerifyWasCalled(pegomock.Once()).
		Delete(pegomock.Any[string](), pegomock.Any[time.Time]()).
		GetCapturedArguments()
	assert.Equal(t, "pr", cPr)
	assert.Equal(t, time.Date(2023, time.January, 2, 15, 04, 05, 0, time.UTC), cTime)
}

func TestLogDelete_FailDate(t *testing.T) {
	initTest(t)
	pegomock.When(logRetrieverMock.Delete(pegomock.Any[string](), pegomock.Any[time.Time]())).ThenReturn(10, nil)
	req := httptest.NewRequest(http.MethodDelete, "/pr/log", nil)
	testCode(t, req, http.StatusBadRequest)
}

func TestLogDelete_FailDB(t *testing.T) {
	initTest(t)
	pegomock.When(logRetrieverMock.Delete(pegomock.Any[string](), pegomock.Any[time.Time]())).ThenReturn(0, fmt.Errorf("err"))
	req := httptest.NewRequest(http.MethodDelete, "/pr/log?to=2023-01-02T15:04:05Z", nil)
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
		LogProvider:      logRetrieverMock,
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
	if !assert.Equal(t, code, resp.Code) {
		assert.Equal(t, code, resp.Code, resp.Body.String())
	}
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
