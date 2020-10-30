package handler

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks/matchers"
	"github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
)

var dbSaverMock *mocks.MockDBSaver

func initLogDBTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	dbSaverMock = mocks.NewMockDBSaver()
}

func TestLogDB(t *testing.T) {
	initLogDBTest(t)
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	ctx.Key = "kkk"
	ctx.Manual = true
	ctx.ResponseCode = 200
	ctx.Value = "value"
	resp := httptest.NewRecorder()
	h := LogDB(newTestHandler(), dbSaverMock).(*logDB)
	h.sync = true

	h.ServeHTTP(resp, req)

	assert.Equal(t, testCode, resp.Code)
	cLog := dbSaverMock.VerifyWasCalledOnce().Save(matchers.AnyPtrToApiLog()).GetCapturedArguments()
	assert.Equal(t, "kkk", cLog.Key)
	assert.Equal(t, 200, cLog.ResponseCode)
	assert.Equal(t, false, cLog.Fail)
	assert.Equal(t, "value", cLog.Value)
	assert.Equal(t, "192.0.2.1", cLog.IP)
	assert.Equal(t, "/duration", cLog.URL)
}

func TestLogDB_NoFail(t *testing.T) {
	initLogDBTest(t)
	req, _ := customContext(httptest.NewRequest("POST", "/duration", nil))
	resp := httptest.NewRecorder()
	h := LogDB(newTestHandler(), dbSaverMock).(*logDB)
	h.sync = true
	pegomock.When(dbSaverMock.Save(matchers.AnyPtrToApiLog())).ThenReturn(errors.New("olia"))

	h.ServeHTTP(resp, req)

	assert.Equal(t, testCode, resp.Code)
}

func TestRespFailCode(t *testing.T) {
	assert.True(t, responseCodeIsFail(100))
	assert.True(t, responseCodeIsFail(400))
	assert.True(t, responseCodeIsFail(500))
	assert.False(t, responseCodeIsFail(200))
	assert.False(t, responseCodeIsFail(299))
}
