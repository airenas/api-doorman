package handler

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/petergtz/pegomock"
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
	ctx.Manual = true
	resp := httptest.NewRecorder()
	pegomock.When(keyValidatorMock.IsValid(pegomock.AnyString(), pegomock.AnyBool())).ThenReturn(true, nil)
	KeyValid(&testHandler{}, keyValidatorMock).ServeHTTP(resp, req)
	assert.Equal(t, testCode, resp.Code)
	cKey, cM := keyValidatorMock.VerifyWasCalledOnce().IsValid(pegomock.AnyString(), pegomock.AnyBool()).GetCapturedArguments()
	assert.Equal(t, "kkk", cKey)
	assert.True(t, cM)
}

func TestKeyValid_Unauthorized(t *testing.T) {
	initKeyValidatorTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	ctx.Manual = true
	resp := httptest.NewRecorder()
	pegomock.When(keyValidatorMock.IsValid(pegomock.AnyString(), pegomock.AnyBool())).ThenReturn(false, nil)
	KeyValid(&testHandler{}, keyValidatorMock).ServeHTTP(resp, req)
	assert.Equal(t, 401, resp.Code)
}

func TestKeyValid_Fail(t *testing.T) {
	initKeyValidatorTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	ctx.Manual = true
	resp := httptest.NewRecorder()
	pegomock.When(keyValidatorMock.IsValid(pegomock.AnyString(), pegomock.AnyBool())).ThenReturn(false, errors.New("olia"))
	KeyValid(&testHandler{}, keyValidatorMock).ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)
}
