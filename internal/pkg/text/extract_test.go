package text

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInit_FailOnWronURL(t *testing.T) {
	_, err := NewExtractor("http://")
	assert.NotNil(t, err)
	_, err = NewExtractor("")
	assert.NotNil(t, err)
}

func TestInit(t *testing.T) {
	d, err := NewExtractor("http://localhost:8000")
	assert.Nil(t, err)
	assert.NotNil(t, d)
	assert.Equal(t, time.Minute, d.timeOut)
}

func initTestServer(t *testing.T, rCode int, body string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(rCode)
		_, _ = rw.Write([]byte(body))
	}))
	return server
}

func TestClient(t *testing.T) {
	var resp textResponse
	resp.Text = "olia"
	rb, _ := json.Marshal(resp)
	server := initTestServer(t, 200, string(rb))
	defer server.Close()
	d, _ := NewExtractor(server.URL)
	d.httpclient = server.Client()

	r, err := d.Get(context.TODO(), "1.wav", strings.NewReader("olia"))

	assert.Nil(t, err)
	assert.Equal(t, "olia", r)
}

func TestClient_TxtFile(t *testing.T) {
	var resp textResponse
	rb, _ := json.Marshal(resp)
	server := initTestServer(t, 200, string(rb))
	defer server.Close()
	d, _ := NewExtractor(server.URL)
	d.httpclient = server.Client()

	r, err := d.Get(context.TODO(), "1.txt", strings.NewReader("olia1"))

	assert.Nil(t, err)
	assert.Equal(t, "olia1", r)
}

func TestClient_PassFile(t *testing.T) {
	var resp textResponse
	resp.Text = "olia"
	rb, _ := json.Marshal(resp)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		_ = req.ParseMultipartForm(32 << 20)
		file, handler, _ := req.FormFile("file")
		defer file.Close()
		assert.Equal(t, "1.wav", handler.Filename)
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, file)
		assert.Equal(t, "olia", buf.String())
		rw.WriteHeader(200)
		_, _ = rw.Write(rb)
	}))
	defer server.Close()
	d, _ := NewExtractor(server.URL)
	d.httpclient = server.Client()

	_, err := d.Get(context.TODO(), "1.wav", strings.NewReader("olia"))

	assert.Nil(t, err)
}

func TestClient_Fail(t *testing.T) {
	var resp textResponse
	resp.Text = "olia"
	rb, _ := json.Marshal(resp)
	server := initTestServer(t, 400, string(rb))
	defer server.Close()
	d, _ := NewExtractor(server.URL)
	d.httpclient = server.Client()

	_, err := d.Get(context.TODO(), "1.wav", strings.NewReader("olia"))

	assert.NotNil(t, err)
}

func TestClient_Timeout(t *testing.T) {
	var resp textResponse
	resp.Text = "olia"
	rb, _ := json.Marshal(resp)
	server := initTestServer(t, 200, string(rb))
	defer server.Close()
	d, _ := NewExtractor(server.URL)
	d.timeOut = time.Duration(0)
	d.httpclient = server.Client()

	_, err := d.Get(context.TODO(), "1.wav", strings.NewReader("olia"))

	assert.NotNil(t, err)
}
