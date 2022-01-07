package cms

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks2"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks2/matchers"
	"github.com/labstack/echo/v4"
	"github.com/petergtz/pegomock"
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
	req := httptest.NewRequest("POST", "/key", mocks.ToReader(api.CreateInput{ID: "1"}))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"kkk"`)
	cVal := prValidarorMock.VerifyWasCalled(pegomock.Once()).Check(pegomock.AnyString()).GetCapturedArguments()
	assert.Equal(t, "pr", cVal)
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
