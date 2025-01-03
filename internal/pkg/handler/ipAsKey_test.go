package handler

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/petergtz/pegomock/v4"
	"github.com/stretchr/testify/assert"

	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
)

var ipSaverMock *mocks.MockIPSaver

func initIPTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	ipSaverMock = mocks.NewMockIPSaver()
}

func TestIP(t *testing.T) {
	initIPTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	resp := httptest.NewRecorder()

	IPAsKey(newTestHandler(), ipSaverMock).ServeHTTP(resp, req)
	_, str := ipSaverMock.VerifyWasCalledOnce().Save(pegomock.Any[context.Context](), pegomock.Any[string]()).GetCapturedArguments()
	assert.Equal(t, 555, resp.Code)
	assert.Equal(t, "192.0.2.1", str)
	assert.Equal(t, "192.0.2.1", ctx.Key)
}

func TestIP_Fail(t *testing.T) {
	initIPTest(t)
	pegomock.When(ipSaverMock.Save(pegomock.Any[context.Context](), pegomock.Any[string]())).ThenReturn("", errors.New("olia"))
	req := httptest.NewRequest("POST", "/duration", nil)
	resp := httptest.NewRecorder()
	IPAsKey(newTestHandler(), ipSaverMock).ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)
}
