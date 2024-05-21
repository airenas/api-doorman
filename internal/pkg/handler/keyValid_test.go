package handler

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/petergtz/pegomock/v4"
	"github.com/stretchr/testify/assert"
)

var keyValidatorMock *mocks.MockKeyValidator

func initKeyValidatorTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	keyValidatorMock = mocks.NewMockKeyValidator()
}

func TestKeyValid(t *testing.T) {
	initKeyValidatorTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	ctx.IP = "1.2.3.4"
	ctx.Manual = true
	resp := httptest.NewRecorder()
	pegomock.When(keyValidatorMock.IsValid(pegomock.Any[string](), pegomock.Any[string](), pegomock.Any[bool]())).
		ThenReturn(true, "id1", []string{"olia"}, nil)
	KeyValid(newTestHandler(), keyValidatorMock).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, []string{"olia"}, ctx.Tags)
	cKey, cIP, cM := keyValidatorMock.VerifyWasCalledOnce().IsValid(pegomock.Any[string](), pegomock.Any[string](), pegomock.Any[bool]()).GetCapturedArguments()
	assert.Equal(t, "kkk", cKey)
	assert.Equal(t, "1.2.3.4", cIP)
	assert.True(t, cM)

	assert.Equal(t, "id1", ctx.KeyID)
}

func TestKeyValid_Unauthorized(t *testing.T) {
	initKeyValidatorTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	ctx.Manual = true
	resp := httptest.NewRecorder()
	pegomock.When(keyValidatorMock.IsValid(pegomock.Any[string](), pegomock.Any[string](), pegomock.Any[bool]())).
		ThenReturn(false, "", nil, nil)
	KeyValid(newTestHandler(), keyValidatorMock).ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)
}

func TestKeyValid_Fail(t *testing.T) {
	initKeyValidatorTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	ctx.Manual = true
	resp := httptest.NewRecorder()
	pegomock.When(keyValidatorMock.IsValid(pegomock.Any[string](), pegomock.Any[string](), pegomock.Any[bool]())).
		ThenReturn(false, "", nil, errors.New("olia"))
	KeyValid(newTestHandler(), keyValidatorMock).ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)
}

func Test_getLimitSetting(t *testing.T) {
	type args struct {
		tags []string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{name: "empty", args: args{tags: []string{}}, want: 0, wantErr: false},
		{name: "parses", args: args{tags: []string{"x-rate-limit:500"}}, want: 500, wantErr: false},
		{name: "several parses", args: args{tags: []string{"olia:100", "x-rate-limit: 500"}}, want: 500, wantErr: false},
		{name: "several parses", args: args{tags: []string{"olia:100", "x-rate-limit: aa500"}}, want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getLimitSetting(tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLimitSetting() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getLimitSetting() = %v, want %v", got, tt.want)
			}
		})
	}
}
