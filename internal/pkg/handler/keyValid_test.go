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
