package cms

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/model"
	"github.com/airenas/api-doorman/internal/pkg/model/permission"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks2"
	"github.com/labstack/echo/v4"
	"github.com/petergtz/pegomock/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var (
	prValidarorMock *mocks.MockPrValidator
	intMock         *mocks2.MockIntegrator

	tData *Data
	tEcho *echo.Echo
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	prValidarorMock = mocks.NewMockPrValidator()
	intMock = mocks2.NewMockIntegrator()
	pegomock.When(prValidarorMock.Check(pegomock.Any[string]())).ThenReturn(true)

	tData = &Data{ProjectValidator: prValidarorMock, Integrator: intMock}
	tEcho = echo.New()

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
	tEcho.Use(authMw)

	InitRoutes(tEcho, tData)
}

func TestWrongPath(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("GET", "/invalid", nil)
	testCode(t, req, 404)
}

func TestAddKey(t *testing.T) {
	initTest(t)
	pegomock.When(intMock.Create(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Any[*api.CreateInput]())).ThenReturn(&api.Key{Key: "kkk"}, true, nil)
	req := httptest.NewRequest("POST", "/key", mocks.ToReader(api.CreateInput{ID: "1", Service: "pr"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp := testCode(t, req, http.StatusCreated)
	bytes, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"kkk"`)
	cVal := prValidarorMock.VerifyWasCalled(pegomock.Once()).Check(pegomock.Any[string]()).GetCapturedArguments()
	assert.Equal(t, "pr", cVal)
}

func TestAddKey_Fail(t *testing.T) {
	type ret struct {
		key api.Key
		ins bool
		err error
	}
	tests := []struct {
		name string
		ret  ret
		inp  io.Reader
		want int
	}{
		{name: "Created", ret: ret{key: api.Key{Key: "kk"}, ins: true, err: nil}, inp: mocks.ToReader(api.CreateInput{ID: "1", Service: "pr"}),
			want: http.StatusCreated},
		{name: "OK", ret: ret{key: api.Key{Key: "kk"}, ins: false, err: nil}, inp: mocks.ToReader(api.CreateInput{ID: "1", Service: "pr"}),
			want: http.StatusConflict},
		{name: "Fail", ret: ret{key: api.Key{Key: "kk"}, ins: false, err: errors.New("olia")},
			inp:  mocks.ToReader(api.CreateInput{ID: "1", Service: "pr"}),
			want: http.StatusInternalServerError},
		{name: "Fail", ret: ret{key: api.Key{Key: "kk"}, ins: false, err: model.NewWrongFieldError("aa", "msg")},
			inp:  mocks.ToReader(api.CreateInput{ID: "1", Service: "pr"}),
			want: http.StatusBadRequest},
		{name: "Fail", ret: ret{key: api.Key{Key: "kk"}, ins: false, err: nil},
			inp:  mocks.ToReader(api.CreateInput{ID: "ID", Service: ""}),
			want: http.StatusBadRequest},
		{name: "Fail", ret: ret{key: api.Key{Key: "kk"}, ins: false, err: nil},
			inp:  strings.NewReader("olia"),
			want: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initTest(t)
			pegomock.When(intMock.Create(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Any[*api.CreateInput]())).
				ThenReturn(&tt.ret.key, tt.ret.ins, tt.ret.err)
			req := httptest.NewRequest(http.MethodPost, "/key", tt.inp)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			testCode(t, req, tt.want)
		})
	}
}

func TestGetKey(t *testing.T) {
	type ret struct {
		key api.Key
		err error
	}
	tests := []struct {
		name string
		ret  ret
		want int
	}{
		{name: "OK", ret: ret{key: api.Key{Key: "kk"}, err: nil}, want: http.StatusOK},
		{name: "Fail", ret: ret{key: api.Key{Key: "kk"}, err: errors.New("olia")},
			want: http.StatusInternalServerError},
		{name: "Fail", ret: ret{key: api.Key{Key: "kk"}, err: model.ErrNoRecord},
			want: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initTest(t)
			pegomock.When(intMock.GetKey(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Any[string]())).
				ThenReturn(&tt.ret.key, tt.ret.err)
			req := httptest.NewRequest(http.MethodGet, "/key/id1", nil)
			testCode(t, req, tt.want)
		})
	}
}

func TestKeyUsage(t *testing.T) {
	type ret struct {
		res api.Usage
		err error
	}
	tests := []struct {
		name   string
		ret    ret
		params map[string]string
		want   int
	}{
		{name: "OK", ret: ret{res: api.Usage{RequestCount: 1}, err: nil}, want: http.StatusOK},
		{name: "Fail", ret: ret{res: api.Usage{RequestCount: 1}, err: errors.New("olia")},
			want: http.StatusInternalServerError},
		{name: "Fail", ret: ret{res: api.Usage{RequestCount: 1}, err: model.ErrNoRecord},
			want: http.StatusBadRequest},
		{name: "From", ret: ret{res: api.Usage{RequestCount: 1}, err: nil},
			params: map[string]string{"from": "2020-01-20T14:50:30Z"},
			want:   http.StatusOK},
		{name: "To", ret: ret{res: api.Usage{RequestCount: 1}, err: nil},
			params: map[string]string{"to": "2020-01-20T14:50:30Z"},
			want:   http.StatusOK},
		{name: "From fail", ret: ret{res: api.Usage{RequestCount: 1}, err: nil},
			params: map[string]string{"from": "xx2020-01-20T14:50:30Z"},
			want:   http.StatusBadRequest},
		{name: "To fail", ret: ret{res: api.Usage{RequestCount: 1}, err: nil},
			params: map[string]string{"to": "xx2020-01-20T14:50:30Z"},
			want:   http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initTest(t)
			pegomock.When(intMock.Usage(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Any[string](), pegomock.Any[*time.Time](),
				pegomock.Any[*time.Time](), pegomock.Any[bool]())).
				ThenReturn(&tt.ret.res, tt.ret.err)
			req := httptest.NewRequest(http.MethodGet, "/key/id1/usage", nil)
			for k, v := range tt.params {
				q := req.URL.Query()
				q.Add(k, v)
				req.URL.RawQuery = q.Encode()
			}
			testCode(t, req, tt.want)
		})
	}
}

func TestKeyUsage_Full(t *testing.T) {
	initTest(t)
	pegomock.When(intMock.Usage(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Any[string](), pegomock.Any[*time.Time](),
		pegomock.Any[*time.Time](), pegomock.Any[bool]())).
		ThenReturn(&api.Usage{RequestCount: 1}, nil)
	req := httptest.NewRequest(http.MethodGet, "/key/id1/usage", nil)
	testCode(t, req, http.StatusOK)
	intMock.VerifyWasCalledOnce().Usage(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Any[string](), pegomock.Any[*time.Time](),
		pegomock.Any[*time.Time](), pegomock.Eq(false))

	req = httptest.NewRequest(http.MethodGet, "/key/id1/usage?full=1", nil)
	testCode(t, req, http.StatusOK)
	intMock.VerifyWasCalledOnce().Usage(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Any[string](), pegomock.Any[*time.Time](),
		pegomock.Any[*time.Time](), pegomock.Eq(true))
}

func TestGetKey_ReturnKey(t *testing.T) {
	initTest(t)
	pegomock.When(intMock.GetKey(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Any[string]())).
		ThenReturn(&api.Key{Key: "aaa", Service: "srv", LastIP: "1.1.1.1"}, nil)
	req := httptest.NewRequest(http.MethodGet, "/key/id1", nil)
	resp := testCode(t, req, http.StatusOK)
	bytes, _ := io.ReadAll(resp.Body)
	var k api.Key
	_ = json.Unmarshal(bytes, &k)
	assert.Equal(t, api.Key{Service: "srv", LastIP: "1.1.1.1"}, k)

	pegomock.When(intMock.GetKey(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Any[string]())).
		ThenReturn(&api.Key{Key: "aaa", Service: "srv", LastIP: "1.1.1.1"}, nil)
	req = httptest.NewRequest(http.MethodGet, "/key/id1?returnKey=1", nil)
	resp = testCode(t, req, http.StatusOK)
	bytes, _ = io.ReadAll(resp.Body)
	_ = json.Unmarshal(bytes, &k)
	assert.Equal(t, api.Key{Key: "aaa", Service: "srv", LastIP: "1.1.1.1"}, k)
}

func testCode(t *testing.T, req *http.Request, code int) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	tEcho.ServeHTTP(resp, req)
	assert.Equal(t, code, resp.Code)
	return resp
}

func Test_validateService(t *testing.T) {
	type args struct {
		project string
		prVRes  bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "OK", args: args{project: "pr", prVRes: true}, wantErr: false},
		{name: "Empty", args: args{project: "", prVRes: true}, wantErr: true},
		{name: "Fail", args: args{project: "pr", prVRes: false}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initTest(t)
			pegomock.When(prValidarorMock.Check(pegomock.Any[string]())).ThenReturn(tt.args.prVRes)
			if err := validateService(tt.args.project, prValidarorMock); (err != nil) != tt.wantErr {
				t.Errorf("validateService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddCredits(t *testing.T) {
	type ret struct {
		res api.Key
		err error
	}
	tests := []struct {
		name string
		ret  ret
		inp  io.Reader
		want int
	}{
		{name: "Created", ret: ret{res: api.Key{Key: "kk"}, err: nil},
			inp:  mocks.ToReader(api.CreditsInput{OperationID: "1", Credits: 100}),
			want: http.StatusOK},
		{name: "Fail", ret: ret{res: api.Key{Key: "kk"}, err: errors.New("olia")},
			inp:  mocks.ToReader(api.CreditsInput{OperationID: "1", Credits: 100}),
			want: http.StatusInternalServerError},
		{name: "Fail", ret: ret{res: api.Key{Key: "kk"}, err: model.NewWrongFieldError("aa", "msg")},
			inp:  mocks.ToReader(api.CreditsInput{OperationID: "1", Credits: 100}),
			want: http.StatusBadRequest},
		{name: "Fail", ret: ret{res: api.Key{Key: "kk"}, err: model.ErrNoRecord},
			inp:  mocks.ToReader(api.CreditsInput{OperationID: "1", Credits: 100}),
			want: http.StatusBadRequest},
		{name: "Fail", ret: ret{res: api.Key{Key: "kk"}, err: errors.New("olia")},
			inp:  strings.NewReader("olia"),
			want: http.StatusBadRequest},
		{name: "Operation exists", ret: ret{res: api.Key{Key: "kk"}, err: model.ErrOperationExists},
			inp:  mocks.ToReader(api.CreditsInput{OperationID: "1", Credits: 100}),
			want: http.StatusConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initTest(t)
			pegomock.When(intMock.AddCredits(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Any[string](), pegomock.Any[*api.CreditsInput]())).
				ThenReturn(&tt.ret.res, tt.ret.err)
			req := httptest.NewRequest(http.MethodPatch, "/key/id/credits", tt.inp)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			testCode(t, req, tt.want)
		})
	}
}

func TestUpdate(t *testing.T) {
	type ret struct {
		res api.Key
		err error
	}
	tests := []struct {
		name string
		ret  ret
		inp  io.Reader
		want int
	}{
		{name: "OK", ret: ret{res: api.Key{Key: "kk"}, err: nil},
			inp:  mocks.ToReader(map[string]interface{}{}),
			want: http.StatusOK},
		{name: "No record", ret: ret{res: api.Key{Key: "kk"}, err: model.ErrNoRecord},
			inp:  mocks.ToReader(map[string]interface{}{}),
			want: http.StatusBadRequest},
		{name: "Field error", ret: ret{res: api.Key{Key: "kk"}, err: model.NewWrongFieldError("", "empty")},
			inp:  mocks.ToReader(map[string]interface{}{}),
			want: http.StatusBadRequest},
		{name: "Fail", ret: ret{res: api.Key{Key: "kk"}, err: errors.New("olia")},
			inp:  mocks.ToReader(map[string]interface{}{}),
			want: http.StatusInternalServerError},
		{name: "Fail wrong input", ret: ret{res: api.Key{Key: "kk"}, err: nil},
			inp:  strings.NewReader("{olia"),
			want: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initTest(t)
			pegomock.When(intMock.Update(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Any[string](), pegomock.Any[*api.UpdateInput]())).
				ThenReturn(&tt.ret.res, tt.ret.err)
			req := httptest.NewRequest(http.MethodPatch, "/key/id", tt.inp)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			testCode(t, req, tt.want)
		})
	}
}

func TestChange(t *testing.T) {
	type ret struct {
		res api.Key
		err error
	}
	tests := []struct {
		name string
		ret  ret
		want int
	}{
		{name: "OK", ret: ret{res: api.Key{Key: "kk"}, err: nil},
			want: http.StatusOK},
		{name: "No record", ret: ret{res: api.Key{Key: "kk"}, err: model.ErrNoRecord},
			want: http.StatusBadRequest},
		{name: "Fail", ret: ret{res: api.Key{Key: "kk"}, err: errors.New("olia")},
			want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initTest(t)
			pegomock.When(intMock.Change(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Any[string]())).ThenReturn(&tt.ret.res, tt.ret.err)
			req := httptest.NewRequest(http.MethodPost, "/key/id/change", nil)
			testCode(t, req, tt.want)
		})
	}
}

func TestKeyGetID(t *testing.T) {
	type ret struct {
		res api.KeyID
		err error
	}
	tests := []struct {
		name string
		ret  ret
		inp  io.Reader
		want int
	}{
		{name: "OK", ret: ret{res: api.KeyID{ID: "kk"}, err: nil},
			inp:  mocks.ToReader(keyByIDInput{Key: "1"}),
			want: http.StatusOK},
		{name: "No key", ret: ret{res: api.KeyID{ID: "kk"}, err: nil},
			inp:  mocks.ToReader(keyByIDInput{Key: ""}),
			want: http.StatusBadRequest},
		{name: "No key", ret: ret{res: api.KeyID{ID: "kk"}, err: nil},
			inp:  strings.NewReader("olia"),
			want: http.StatusBadRequest},
		{name: "No key", ret: ret{res: api.KeyID{ID: "kk"}, err: errors.New("olia")},
			inp:  mocks.ToReader(keyByIDInput{Key: "1"}),
			want: http.StatusInternalServerError},
		{name: "No key", ret: ret{res: api.KeyID{ID: "kk"}, err: model.ErrNoRecord},
			inp:  mocks.ToReader(keyByIDInput{Key: "1"}),
			want: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initTest(t)
			pegomock.When(intMock.GetKeyID(pegomock.Any[context.Context](), pegomock.Any[*model.User](), pegomock.Any[string]())).
				ThenReturn(&tt.ret.res, tt.ret.err)
			req := httptest.NewRequest(http.MethodPost, "/keyID", tt.inp)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			testCode(t, req, tt.want)
		})
	}
}
