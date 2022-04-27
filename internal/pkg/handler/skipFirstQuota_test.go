package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
)

var getCountMock *mocks.MockCountGetter

func initSkipFirstQuotaTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	getCountMock = mocks.NewMockCountGetter()
}

func TestSkipFirstQuota_Clears(t *testing.T) {
	initSkipFirstQuotaTest(t)
	pegomock.When(getCountMock.GetParamName()).ThenReturn("olia")
	pegomock.When(getCountMock.Get(pegomock.AnyString())).ThenReturn(int64(0), nil)

	req := httptest.NewRequest("POST", "/text?olia=111", nil)
	req, ctx := customContext(req)
	ctx.QuotaValue = 10
	resp := httptest.NewRecorder()

	SkipFirstQuota(newTestHandler(), getCountMock).ServeHTTP(resp, req)
	ci := getCountMock.VerifyWasCalledOnce().Get(pegomock.AnyString()).GetCapturedArguments()
	assert.Equal(t, 555, resp.Code)
	assert.Equal(t, "111", ci)
	assert.InDelta(t, 0, ctx.QuotaValue, 0.00001)
}

func TestSkipFirstQuota_Leaves(t *testing.T) {
	initSkipFirstQuotaTest(t)
	pegomock.When(getCountMock.GetParamName()).ThenReturn("olia")
	pegomock.When(getCountMock.Get(pegomock.AnyString())).ThenReturn(int64(1), nil)

	req := httptest.NewRequest("POST", "/text?olia=111", nil)
	req, ctx := customContext(req)
	ctx.QuotaValue = 10
	resp := httptest.NewRecorder()

	SkipFirstQuota(newTestHandler(), getCountMock).ServeHTTP(resp, req)
	assert.Equal(t, 555, resp.Code)
	assert.InDelta(t, 10, ctx.QuotaValue, 0.00001)
}

func TestSkipFirstQuota_Fails_NoParam(t *testing.T) {
	initSkipFirstQuotaTest(t)
	pegomock.When(getCountMock.GetParamName()).ThenReturn("olia1")
	pegomock.When(getCountMock.Get(pegomock.AnyString())).ThenReturn(int64(1), nil)

	req := httptest.NewRequest("POST", "/text?olia=111", nil)
	req, ctx := customContext(req)
	ctx.QuotaValue = 10
	resp := httptest.NewRecorder()

	SkipFirstQuota(newTestHandler(), getCountMock).ServeHTTP(resp, req)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestSkipFirstQuota_Fails_Service(t *testing.T) {
	initSkipFirstQuotaTest(t)
	pegomock.When(getCountMock.GetParamName()).ThenReturn("olia")
	pegomock.When(getCountMock.Get(pegomock.AnyString())).ThenReturn(int64(0), errors.New("olia"))

	req := httptest.NewRequest("POST", "/text?olia=111", nil)
	req, ctx := customContext(req)
	ctx.QuotaValue = 10
	resp := httptest.NewRecorder()

	SkipFirstQuota(newTestHandler(), getCountMock).ServeHTTP(resp, req)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}
