package admin

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/model"
	"github.com/airenas/api-doorman/internal/pkg/model/permission"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/labstack/echo/v4"
	"github.com/petergtz/pegomock/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

var (
	oneKeyRetrieverMock *mocks.MockOneKeyRetriever
	logRetrieverMock    *mocks.MockLogProvider
	prValidarorMock     *mocks.MockPrValidator
	uRestorer           *mocks.MockUsageRestorer

	tData *Data
	tEcho *echo.Echo
	tResp *httptest.ResponseRecorder
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	oneKeyRetrieverMock = mocks.NewMockOneKeyRetriever()
	logRetrieverMock = mocks.NewMockLogProvider()
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

func TestKey(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Eq("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	req := httptest.NewRequest("GET", "/pr/key/kkk", nil)
	resp := testCode(t, req, 200)
	bytes, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"kkk"`)
	cVal := prValidarorMock.VerifyWasCalled(pegomock.Once()).Check(pegomock.Any[string]()).GetCapturedArguments()
	assert.Equal(t, "pr", cVal)
}

func TestKey_ReturnsFull(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Eq("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	pegomock.When(logRetrieverMock.GetLogs(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Eq("kkk"))).ThenReturn([]*adminapi.Log{{IP: "101010"}}, nil)
	req := httptest.NewRequest("GET", "/pr/key/kkk?full=1", nil)
	resp := testCode(t, req, 200)
	bytes, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"ip":"101010"`)
}

func TestKey_ReturnsFullByKeyID(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Eq("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk", ID: "olia"}, nil)
	pegomock.When(logRetrieverMock.GetLogs(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Eq("kkk"))).ThenReturn([]*adminapi.Log{{IP: "101010"}}, nil)
	pegomock.When(logRetrieverMock.GetLogs(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Eq("olia"))).ThenReturn([]*adminapi.Log{{IP: "101011"}}, nil)
	req := httptest.NewRequest("GET", "/pr/key/kkk?full=1", nil)
	resp := testCode(t, req, 200)
	bytes, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"ip":"101010"`)
}

func TestKey_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Eq("kkk"))).ThenReturn(nil, errors.New("fail"))
	req := httptest.NewRequest("GET", "/pr/key/kkk", nil)
	testCode(t, req, 500)
}

func TestKey_FailIDCall(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Eq("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk", ID: "olia"}, nil)
	pegomock.When(logRetrieverMock.GetLogs(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Eq("kkk"))).ThenReturn(nil, errors.New("fail"))
	req := httptest.NewRequest("GET", "/pr/key/kkk?full=1", nil)
	testCode(t, req, 500)
}

func TestKey_FailFull(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Eq("kkk"))).ThenReturn(&adminapi.Key{Key: "kkk"}, nil)
	pegomock.When(logRetrieverMock.GetLogs(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Eq("kkk"))).ThenReturn(nil, errors.New("fail"))
	req := httptest.NewRequest("GET", "/pr/key/kkk?full=1", nil)
	testCode(t, req, 500)
}

func TestKey_FailNoKey(t *testing.T) {
	initTest(t)
	pegomock.When(oneKeyRetrieverMock.Get(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Eq("kkk"))).ThenReturn(nil, model.ErrNoRecord)
	req := httptest.NewRequest("GET", "/pr/key/kkk", nil)
	testCode(t, req, 400)
}

func TestRestore(t *testing.T) {
	initTest(t)
	pegomock.When(uRestorer.RestoreUsage(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[bool](), pegomock.Any[string](), pegomock.Any[string]())).
		ThenReturn(nil)
	req := httptest.NewRequest(http.MethodPost, "/pr/restore/m:rID", toReader(restoreReq{Error: "err msg"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	_ = testCode(t, req, http.StatusOK)
	_, cPr, cManual, cReq, cErr := uRestorer.VerifyWasCalled(pegomock.Once()).
		RestoreUsage(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[bool](), pegomock.Any[string](), pegomock.Any[string]()).
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
	pegomock.When(uRestorer.RestoreUsage(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[bool](), pegomock.Any[string](), pegomock.Any[string]())).
		ThenReturn(model.ErrNoRecord)
	testCode(t, req, http.StatusBadRequest)

	req = httptest.NewRequest(http.MethodPost, "/pr/restore/m:rID", toReader(restoreReq{Error: "err msg"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	pegomock.When(uRestorer.RestoreUsage(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[bool](), pegomock.Any[string](), pegomock.Any[string]())).
		ThenReturn(model.ErrLogRestored)
	testCode(t, req, http.StatusConflict)

	req = httptest.NewRequest(http.MethodPost, "/pr/restore/m:rID", toReader(restoreReq{Error: "err msg"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	pegomock.When(uRestorer.RestoreUsage(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[bool](), pegomock.Any[string](), pegomock.Any[string]())).
		ThenReturn(errors.New("olia"))
	testCode(t, req, http.StatusInternalServerError)
}

// func TestLogList(t *testing.T) {
// 	initTest(t)
// 	pegomock.When(logRetrieverMock.ListLogs(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[time.Time]())).
// 		ThenReturn([]*adminapi.Log{{KeyID: "1", ResponseCode: 200}, {KeyID: "2", ResponseCode: 400}}, nil)
// 	req := httptest.NewRequest(http.MethodGet, "/pr/log?to=2023-01-02T15:04:05Z", nil)
// 	resp := testCode(t, req, http.StatusOK)
// 	bytes, _ := io.ReadAll(resp.Body)
// 	assert.Contains(t, string(bytes), `"keyID":"1"`)
// 	assert.Contains(t, string(bytes), `"keyID":"2"`)
// 	assert.Contains(t, string(bytes), `"response":400`)
// 	_, cPr, cTime := logRetrieverMock.VerifyWasCalled(pegomock.Once()).
// 		ListLogs(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[time.Time]()).
// 		GetCapturedArguments()
// 	assert.Equal(t, "pr", cPr)
// 	assert.Equal(t, time.Date(2023, time.January, 2, 15, 04, 05, 0, time.UTC), cTime)
// }

// func TestLogList_FailDate(t *testing.T) {
// 	initTest(t)
// 	pegomock.When(logRetrieverMock.ListLogs(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[time.Time]())).
// 		ThenReturn([]*adminapi.Log{{KeyID: "1", ResponseCode: 200}, {KeyID: "2", ResponseCode: 400}}, nil)
// 	req := httptest.NewRequest(http.MethodGet, "/pr/log", nil)
// 	testCode(t, req, http.StatusBadRequest)
// }

// func TestLogList_FailDB(t *testing.T) {
// 	initTest(t)
// 	pegomock.When(logRetrieverMock.ListLogs(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[time.Time]())).
// 		ThenReturn(nil, fmt.Errorf("err"))
// 	req := httptest.NewRequest(http.MethodGet, "/pr/log?to=2023-01-02T15:04:05Z", nil)
// 	testCode(t, req, http.StatusInternalServerError)
// }

// func TestLogDelete(t *testing.T) {
// 	initTest(t)
// 	pegomock.When(logRetrieverMock.DeleteLogs(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[time.Time]())).ThenReturn(10, nil)
// 	req := httptest.NewRequest(http.MethodDelete, "/pr/log?to=2023-01-02T15:04:05Z", nil)
// 	resp := testCode(t, req, http.StatusOK)
// 	bytes, _ := io.ReadAll(resp.Body)
// 	assert.Equal(t, `{"deleted":10}`, string(bytes))
// 	_, cPr, cTime := logRetrieverMock.VerifyWasCalled(pegomock.Once()).
// 		DeleteLogs(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[time.Time]()).
// 		GetCapturedArguments()
// 	assert.Equal(t, "pr", cPr)
// 	assert.Equal(t, time.Date(2023, time.January, 2, 15, 04, 05, 0, time.UTC), cTime)
// }

// func TestLogDelete_FailDate(t *testing.T) {
// 	initTest(t)
// 	pegomock.When(logRetrieverMock.DeleteLogs(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[time.Time]())).ThenReturn(10, nil)
// 	req := httptest.NewRequest(http.MethodDelete, "/pr/log", nil)
// 	testCode(t, req, http.StatusBadRequest)
// }

// func TestLogDelete_FailDB(t *testing.T) {
// 	initTest(t)
// 	pegomock.When(logRetrieverMock.DeleteLogs(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[time.Time]())).ThenReturn(0, fmt.Errorf("err"))
// 	req := httptest.NewRequest(http.MethodDelete, "/pr/log?to=2023-01-02T15:04:05Z", nil)
// 	testCode(t, req, http.StatusInternalServerError)
// }

func toReader(key interface{}) io.Reader {
	bytes, _ := json.Marshal(key)
	return strings.NewReader(string(bytes))
}

func newTestData() *Data {
	authMw := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()
			ctx := r.Context()
			user := &model.User{Name: "olia",
				Permissions: map[permission.Enum]bool{permission.Everything: true},
				Projects:    []string{"test"},
				MaxLimit:    1000,
				MaxValidTo:  time.Now().AddDate(1, 0, 0),
			}
			c.SetRequest(r.WithContext(context.WithValue(ctx, model.CtxUser, user)))
			return next(c)
		}
	}
	res := &Data{
		ProjectValidator: prValidarorMock,
		UsageRestorer:    uRestorer,
		Auth:             authMw,
		OneKeyGetter:     oneKeyRetrieverMock,
		LogProvider:      logRetrieverMock,
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
