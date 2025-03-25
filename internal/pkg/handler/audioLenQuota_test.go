package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/petergtz/pegomock/v4"
	"github.com/stretchr/testify/assert"

	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
)

var audioLenGetterMock *mocks.MockAudioLenGetter

func initAudioTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	audioLenGetterMock = mocks.NewMockAudioLenGetter()
}

func TestAudio(t *testing.T) {
	initAudioTest(t)
	req := newTestAudioRequest("test.mp3")
	resp := httptest.NewRecorder()

	AudioLenQuota(newTestHandler(), "file", audioLenGetterMock).ServeHTTP(resp, req)
	_, cf, _ := audioLenGetterMock.VerifyWasCalledOnce().Get(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[io.Reader]()).GetCapturedArguments()
	assert.Equal(t, 555, resp.Code)
	assert.Equal(t, "test.mp3", cf)
}

func TestAudio_Fail(t *testing.T) {
	initAudioTest(t)
	req := newTestAudioRequest("test.mp3")
	resp := httptest.NewRecorder()

	AudioLenQuota(newTestHandler(), "file1", audioLenGetterMock).ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)
}

func TestAudio_FailBody(t *testing.T) {
	initAudioTest(t)
	req := httptest.NewRequest("POST", "/duration", nil)
	resp := httptest.NewRecorder()
	AudioLenQuota(newTestHandler(), "file", audioLenGetterMock).ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)
}

func TestAudio_FailAudio(t *testing.T) {
	initAudioTest(t)
	req := newTestAudioRequest("test.mp3")
	resp := httptest.NewRecorder()
	pegomock.When(audioLenGetterMock.Get(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[io.Reader]())).ThenReturn(0.0, errors.New("olia"))
	AudioLenQuota(newTestHandler(), "file", audioLenGetterMock).ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)
}

func TestAudio_SetResult(t *testing.T) {
	initAudioTest(t)
	req, ctx := customContext(newTestAudioRequest("test.mp3"))
	resp := httptest.NewRecorder()
	pegomock.When(audioLenGetterMock.Get(pegomock.Any[context.Context](), pegomock.Any[string](), pegomock.Any[io.Reader]())).ThenReturn(10.0, nil)
	AudioLenQuota(newTestHandler(), "file", audioLenGetterMock).ServeHTTP(resp, req)
	assert.Equal(t, 555, resp.Code)
	assert.Equal(t, 10.0, ctx.QuotaValue)
}

func newTestAudioRequest(file string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if file != "" {
		part, _ := writer.CreateFormFile("file", file)
		_, _ = io.Copy(part, strings.NewReader("body"))
	}
	writer.Close()
	req := httptest.NewRequest("POST", "/duration", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

type testHandler struct {
	code int
	r    *http.Request
}

func newTestHandlerWithCode(code int) *testHandler {
	return &testHandler{code: code}
}

func newTestHandler() *testHandler {
	return &testHandler{code: testCode}
}

const testCode = 555

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(h.code)
	_, ctx := customContext(r)
	ctx.ResponseCode = h.code
	h.r = r
}
