package handler

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/petergtz/pegomock/v4"
	"github.com/stretchr/testify/assert"
)

var getTextMock *mocks.MockTextGetter

func initToTextAndQuotaTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	getTextMock = mocks.NewMockTextGetter()
}

func TestToTextAndQuotaTest(t *testing.T) {
	initToTextAndQuotaTest(t)
	req := newToTextAndQuotaTestRequest("test.epub", "text olia")
	req, ctx := customContext(req)
	resp := httptest.NewRecorder()

	pegomock.When(getTextMock.Get(pegomock.Any[string](), pegomock.Any[io.Reader]())).ThenReturn("olia olia", nil)
	ToTextAndQuota(newTestHandler(), "file", getTextMock).ServeHTTP(resp, req)
	cf, _ := getTextMock.VerifyWasCalledOnce().Get(pegomock.Any[string](), pegomock.Any[io.Reader]()).GetCapturedArguments()
	assert.Equal(t, 555, resp.Code)
	assert.Equal(t, "test.epub", cf)
	assert.InDelta(t, 9, ctx.QuotaValue, 0.00001)
}

func TestToTextAndQuotaTest_FailService(t *testing.T) {
	initToTextAndQuotaTest(t)
	req := newToTextAndQuotaTestRequest("test.epub", "text olia")
	resp := httptest.NewRecorder()

	pegomock.When(getTextMock.Get(pegomock.Any[string](), pegomock.Any[io.Reader]())).ThenReturn("", errors.New("olia error"))
	ToTextAndQuota(newTestHandler(), "file", getTextMock).ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)
}

func TestToTextAndQuotaTest_FailInput(t *testing.T) {
	initToTextAndQuotaTest(t)
	req := newToTextAndQuotaTestRequest("test.epub", "text olia")
	resp := httptest.NewRecorder()

	ToTextAndQuota(newTestHandler(), "file1", getTextMock).ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)
}

func TestToTextAndQuotaTest_FailBodyInput(t *testing.T) {
	initToTextAndQuotaTest(t)
	req := httptest.NewRequest("POST", "/text", nil)
	resp := httptest.NewRecorder()

	ToTextAndQuota(newTestHandler(), "file", getTextMock).ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)
}

func TestToTextAndQuotaTest_PassFormValues(t *testing.T) {
	initToTextAndQuotaTest(t)
	req := newToTextAndQuotaTestRequest("test.epub", "text olia")
	resp := httptest.NewRecorder()

	th := newTestHandler()
	ToTextAndQuota(th, "file", getTextMock).ServeHTTP(resp, req)
	assert.Equal(t, 555, resp.Code)
	assert.Equal(t, "olia@olia.eu", th.r.FormValue("email"))
	assert.Equal(t, []string{"olia1", "olia2"}, th.r.MultipartForm.Value["olia"])
}

func TestToTextAndQuotaTest_ContentLength(t *testing.T) {
	initToTextAndQuotaTest(t)
	req := newToTextAndQuotaTestRequest("test.epub", "text olia")
	resp := httptest.NewRecorder()

	pegomock.When(getTextMock.Get(pegomock.Any[string](), pegomock.Any[io.Reader]())).ThenReturn("olia olia -- 123", nil)

	th := newTestHandler()
	ToTextAndQuota(th, "file", getTextMock).ServeHTTP(resp, req)
	assert.Equal(t, "618", th.r.Header.Get("Content-Length"))
	assert.Equal(t, int64(618), th.r.ContentLength)
}

func TestToTextAndQuotaTest_passTxtFile(t *testing.T) {
	initToTextAndQuotaTest(t)
	req := newToTextAndQuotaTestRequest("test.epub", "text olia")
	resp := httptest.NewRecorder()

	ts := `"olia olia"\ntada`
	pegomock.When(getTextMock.Get(pegomock.Any[string](), pegomock.Any[io.Reader]())).ThenReturn(ts, nil)

	th := newTestHandler()
	ToTextAndQuota(th, "file", getTextMock).ServeHTTP(resp, req)
	file, _, err := th.r.FormFile("file")
	assert.Nil(t, err)
	fd, _ := io.ReadAll(file)
	assert.Equal(t, ts, string(fd))
}

func newToTextAndQuotaTestRequest(file, bodyText string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if file != "" {
		part, _ := writer.CreateFormFile("file", file)
		_, _ = io.Copy(part, strings.NewReader(bodyText))
	}
	_ = writer.WriteField("olia", "olia1")
	_ = writer.WriteField("olia", "olia2")
	_ = writer.WriteField("email", "olia@olia.eu")
	writer.Close()
	req := httptest.NewRequest("POST", "/text", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func Test_toTxtExt(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "txt", args: args{s: "file.txt"}, want: "file.txt"},
		{name: "epub", args: args{s: "file.epub"}, want: "file.txt"},
		{name: "with dir", args: args{s: "olia/file.pdf"}, want: "olia/file.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toTxtExt(tt.args.s); got != tt.want {
				t.Errorf("toTxtExt() = %v, want %v", got, tt.want)
			}
		})
	}
}
