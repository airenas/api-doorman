package handler

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
)

var quotaValidatorMock *mocks.MockQuotaValidator

func inituotaValidateTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	quotaValidatorMock = mocks.NewMockQuotaValidator()
}

func TestQuotaValidate(t *testing.T) {
	inituotaValidateTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	ctx.QuotaValue = 100
	resp := httptest.NewRecorder()
	pegomock.When(quotaValidatorMock.SaveValidate(pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyFloat64())).ThenReturn(true, 10.0, 20.0, nil)

	QuotaValidate(newTestHandler(), quotaValidatorMock).ServeHTTP(resp, req)

	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, 0, ctx.ResponseCode)
	cKey, _, cQuota := quotaValidatorMock.VerifyWasCalledOnce().SaveValidate(pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyFloat64()).
		GetCapturedArguments()
	assert.Equal(t, "kkk", cKey)
	assert.Equal(t, 100.0, cQuota)
}

func TestQuotaValidate_Header(t *testing.T) {
	inituotaValidateTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	ctx.QuotaValue = 100
	resp := httptest.NewRecorder()
	pegomock.When(quotaValidatorMock.SaveValidate(pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyFloat64())).ThenReturn(true, 10.0, 20.0, nil)

	QuotaValidate(newTestHandler(), quotaValidatorMock).ServeHTTP(resp, req)
	assert.Equal(t, "10", resp.Header().Get("X-Rate-Limit-Remaining"))
	assert.Equal(t, "20", resp.Header().Get("X-Rate-Limit-Limit"))
}

func TestQuotaValidate_Fail(t *testing.T) {
	inituotaValidateTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	ctx.QuotaValue = 100
	resp := httptest.NewRecorder()
	pegomock.When(quotaValidatorMock.SaveValidate(pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyFloat64())).
		ThenReturn(true, 10.0, 20.0, errors.New("olia"))

	QuotaValidate(newTestHandler(), quotaValidatorMock).ServeHTTP(resp, req)

	assert.Equal(t, 500, resp.Code)
	assert.Equal(t, 500, ctx.ResponseCode)
}

func TestQuotaValidate_Unauthorized(t *testing.T) {
	inituotaValidateTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	ctx.QuotaValue = 100
	resp := httptest.NewRecorder()
	pegomock.When(quotaValidatorMock.SaveValidate(pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyFloat64())).
		ThenReturn(false, 10.0, 20.0, nil)

	QuotaValidate(newTestHandler(), quotaValidatorMock).ServeHTTP(resp, req)

	assert.Equal(t, 403, resp.Code)
	assert.Equal(t, 403, ctx.ResponseCode)
}
