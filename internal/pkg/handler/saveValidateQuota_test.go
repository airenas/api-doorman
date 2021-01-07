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
	ctx.Manual = true
	resp := httptest.NewRecorder()
	pegomock.When(quotaValidatorMock.SaveValidate(pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyBool(),
		pegomock.AnyFloat64())).ThenReturn(true, 10.0, 20.0, nil)

	QuotaValidate(newTestHandler(), quotaValidatorMock).ServeHTTP(resp, req)

	assert.Equal(t, testCode, resp.Code)
	assert.Equal(t, testCode, ctx.ResponseCode)
	cKey, _, cManual, cQuota := quotaValidatorMock.VerifyWasCalledOnce().SaveValidate(pegomock.AnyString(), pegomock.AnyString(),
		pegomock.AnyBool(), pegomock.AnyFloat64()).GetCapturedArguments()
	assert.Equal(t, "kkk", cKey)
	assert.Equal(t, 100.0, cQuota)
	assert.True(t, cManual)
}

func TestQuotaValidate_Header(t *testing.T) {
	inituotaValidateTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	ctx.QuotaValue = 100
	resp := httptest.NewRecorder()
	pegomock.When(quotaValidatorMock.SaveValidate(pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyBool(),
		pegomock.AnyFloat64())).ThenReturn(true, 10.0, 20.0, nil)

	QuotaValidate(newTestHandlerWithCode(200), quotaValidatorMock).ServeHTTP(resp, req)
	assert.Equal(t, "10", resp.Header().Get("X-Rate-Limit-Remaining"))
	assert.Equal(t, "20", resp.Header().Get("X-Rate-Limit-Limit"))
}

func TestQuotaValidate_Fail(t *testing.T) {
	inituotaValidateTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	ctx.QuotaValue = 100
	resp := httptest.NewRecorder()
	pegomock.When(quotaValidatorMock.SaveValidate(pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyBool(),
		pegomock.AnyFloat64())).ThenReturn(true, 10.0, 20.0, errors.New("olia"))

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
	pegomock.When(quotaValidatorMock.SaveValidate(pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyBool(),
		pegomock.AnyFloat64())).ThenReturn(false, 10.0, 20.0, nil)

	QuotaValidate(newTestHandler(), quotaValidatorMock).ServeHTTP(resp, req)

	assert.Equal(t, 403, resp.Code)
	assert.Equal(t, 403, ctx.ResponseCode)
}

func TestQuotaValidate_NoRestore(t *testing.T) {
	inituotaValidateTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	ctx.QuotaValue = 100
	resp := httptest.NewRecorder()
	pegomock.When(quotaValidatorMock.SaveValidate(pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyBool(),
		pegomock.AnyFloat64())).ThenReturn(true, 10.0, 20.0, nil)

	QuotaValidate(newTestHandlerWithCode(200), quotaValidatorMock).ServeHTTP(resp, req)

	quotaValidatorMock.VerifyWasCalled(pegomock.Never()).Restore(pegomock.AnyString(), pegomock.AnyBool(), pegomock.AnyFloat64())
}

func TestQuotaValidate_Restore(t *testing.T) {
	inituotaValidateTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	ctx.QuotaValue = 100
	ctx.Manual = true
	resp := httptest.NewRecorder()
	pegomock.When(quotaValidatorMock.SaveValidate(pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyBool(),
		pegomock.AnyFloat64())).ThenReturn(true, 10.0, 20.0, nil)
	pegomock.When(quotaValidatorMock.Restore(pegomock.AnyString(), pegomock.AnyBool(),
		pegomock.AnyFloat64())).ThenReturn(5.0, 25.0, nil)

	QuotaValidate(newTestHandlerWithCode(503), quotaValidatorMock).ServeHTTP(resp, req)

	cKey, cManual, cQuota := quotaValidatorMock.VerifyWasCalled(pegomock.Once()).Restore(pegomock.AnyString(), pegomock.AnyBool(), pegomock.AnyFloat64()).
		GetCapturedArguments()
	assert.Equal(t, "kkk", cKey)
	assert.Equal(t, 100.0, cQuota)
	assert.True(t, cManual)
	assert.Equal(t, "5", resp.Header().Get("X-Rate-Limit-Remaining"))
	assert.Equal(t, "25", resp.Header().Get("X-Rate-Limit-Limit"))
}

func TestQuotaValidate_RestoreFail(t *testing.T) {
	inituotaValidateTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	ctx.QuotaValue = 100
	ctx.Manual = true
	resp := httptest.NewRecorder()
	pegomock.When(quotaValidatorMock.SaveValidate(pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyBool(),
		pegomock.AnyFloat64())).ThenReturn(true, 10.0, 20.0, nil)
	pegomock.When(quotaValidatorMock.Restore(pegomock.AnyString(), pegomock.AnyBool(),
		pegomock.AnyFloat64())).ThenReturn(0.0, 0.0, errors.New("olia"))

	QuotaValidate(newTestHandlerWithCode(404), quotaValidatorMock).ServeHTTP(resp, req)

	quotaValidatorMock.VerifyWasCalled(pegomock.Once()).Restore(pegomock.AnyString(), pegomock.AnyBool(), pegomock.AnyFloat64())
	assert.Equal(t, 404, resp.Code)
}

func TestIsServiceFailure(t *testing.T) {
	assert.True(t, isServiceFailure(404))
	assert.True(t, isServiceFailure(500))
	assert.True(t, isServiceFailure(501))
	assert.True(t, isServiceFailure(503))
	assert.False(t, isServiceFailure(200))
	assert.False(t, isServiceFailure(202))
	assert.False(t, isServiceFailure(400))
	assert.False(t, isServiceFailure(403))
}
