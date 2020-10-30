package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxy_Response(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(442)
	}))
	defer server.Close()
	req, ctx := customContext(httptest.NewRequest("POST", "/duration", nil))
	resp := httptest.NewRecorder()

	surl, _ := url.Parse(server.URL)
	Proxy(surl).ServeHTTP(resp, req)
	assert.Equal(t, 442, resp.Code)
	assert.Equal(t, 442, ctx.ResponseCode)
}
