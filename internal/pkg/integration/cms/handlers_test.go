package cms

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks2"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks2/matchers"
	"github.com/labstack/echo/v4"
	"github.com/petergtz/pegomock"
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
	pegomock.When(prValidarorMock.Check(pegomock.AnyString())).ThenReturn(true)

	tData = &Data{ProjectValidator: prValidarorMock, Integrator: intMock}
	tEcho = echo.New()
	InitRoutes(tEcho, tData)
}

func TestWrongPath(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("GET", "/invalid", nil)
	testCode(t, req, 404)
}

func TestAddKey(t *testing.T) {
	initTest(t)
	pegomock.When(intMock.Create(matchers.AnyPtrToApiCreateInput())).ThenReturn(&api.Key{Key: "kkk"}, true, nil)
	req := httptest.NewRequest("POST", "/key", mocks.ToReader(api.CreateInput{ID: "1", Service: "pr"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp := testCode(t, req, http.StatusCreated)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"kkk"`)
	cVal := prValidarorMock.VerifyWasCalled(pegomock.Once()).Check(pegomock.AnyString()).GetCapturedArguments()
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
			want: http.StatusOK},
		{name: "Fail", ret: ret{key: api.Key{Key: "kk"}, ins: false, err: errors.New("olia")},
			inp:  mocks.ToReader(api.CreateInput{ID: "1", Service: "pr"}),
			want: http.StatusInternalServerError},
		{name: "Fail", ret: ret{key: api.Key{Key: "kk"}, ins: false, err: &api.ErrField{Field: "aa", Msg: "msg"}},
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
			pegomock.When(intMock.Create(matchers.AnyPtrToApiCreateInput())).
				ThenReturn(&tt.ret.key, tt.ret.ins, tt.ret.err)
			req := httptest.NewRequest(http.MethodPost, "/key", tt.inp)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			testCode(t, req, tt.want)
		})
	}
}

func TestGetKey_Fail(t *testing.T) {
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
		{name: "Fail", ret: ret{key: api.Key{Key: "kk"}, err: api.ErrNoRecord},
			want: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initTest(t)
			pegomock.When(intMock.GetKey(pegomock.AnyString())).
				ThenReturn(&tt.ret.key, tt.ret.err)
			req := httptest.NewRequest(http.MethodGet, "/key/id1", nil)
			testCode(t, req, tt.want)
		})
	}
}

func TestGetKey_ReturnKey(t *testing.T) {
	initTest(t)
	pegomock.When(intMock.GetKey(pegomock.AnyString())).
		ThenReturn(&api.Key{Key: "aaa", Service: "srv", LastIP: "1.1.1.1"}, nil)
	req := httptest.NewRequest(http.MethodGet, "/key/id1", nil)
	resp := testCode(t, req, http.StatusOK)
	bytes, _ := ioutil.ReadAll(resp.Body)
	var k api.Key
	json.Unmarshal(bytes, &k)
	assert.Equal(t, api.Key{Service: "srv", LastIP: "1.1.1.1"}, k)
	
	pegomock.When(intMock.GetKey(pegomock.AnyString())).
		ThenReturn(&api.Key{Key: "aaa", Service: "srv", LastIP: "1.1.1.1"}, nil)
	req = httptest.NewRequest(http.MethodGet, "/key/id1?returnKey=1", nil)
	resp = testCode(t, req, http.StatusOK)
	bytes, _ = ioutil.ReadAll(resp.Body)
	json.Unmarshal(bytes, &k)
	assert.Equal(t, api.Key{Key: "aaa", Service: "srv", LastIP: "1.1.1.1"}, k)
}

func Test_parseDateParam(t *testing.T) {
	type args struct {
		s string
	}
	wanted, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{name: "Empty", args: args{s: ""}, wantErr: false},
		{name: "Error", args: args{s: "err"}, wantErr: true},
		{name: "Error", args: args{s: "2006-13-02T15:04:05Z"}, wantErr: true},
		{name: "Error", args: args{s: "2006-11-31T15:04:05Z"}, wantErr: true},
		{name: "Parse", args: args{s: "2006-01-02T15:04:05Z"}, want: wanted, wantErr: false},
		{name: "Parse TZ", args: args{s: "2006-01-02T16:04:05+01:00"}, want: wanted, wantErr: false},
		{name: "Parse TZ", args: args{s: "2006-01-02T17:04:05+02:00"}, want: wanted, wantErr: false},
		{name: "Parse TZ", args: args{s: "2006-01-02T12:04:05-03:00"}, want: wanted, wantErr: false},
		{name: "Parse TZ", args: args{s: "2006-01-02T11:34:05-03:30"}, want: wanted, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDateParam(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDateParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.want.IsZero() {
				if got == nil || !(got.Before(tt.want.Add(time.Millisecond)) &&
					got.After(tt.want.Add(-time.Millisecond))) {
					t.Errorf("parseDateParam() = %v, want %v", got, tt.want)
				}
			}
		})
	}
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
			pegomock.When(prValidarorMock.Check(pegomock.AnyString())).ThenReturn(tt.args.prVRes)
			if err := validateService(tt.args.project, prValidarorMock); (err != nil) != tt.wantErr {
				t.Errorf("validateService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
