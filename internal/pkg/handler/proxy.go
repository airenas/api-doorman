package handler

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/airenas/go-app/pkg/goapp"
)

type proxy struct {
	url *url.URL
}

//Proxy creates handler
func Proxy(url *url.URL) http.Handler {
	res := &proxy{}
	res.url = url
	return res
}

func (h *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	rn.URL.Host = h.url.Host
	rn.URL.Scheme = h.url.Scheme
	rn.Header.Set("X-Forwarded-Host", rn.Header.Get("Host"))
	rn.Host = h.url.Host
	proxy := httputil.NewSingleHostReverseProxy(h.url)
	proxy.ModifyResponse = func(resp *http.Response) (err error) {
		ctx.ResponseCode = resp.StatusCode
		return nil
	}
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		goapp.Log.Errorf("http: proxy error: %v", err)
		rw.WriteHeader(http.StatusBadGateway)
		ctx.ResponseCode = http.StatusBadGateway
	}
	proxy.ServeHTTP(w, rn)
}

func (h *proxy) Info(pr string) string {
	return pr + fmt.Sprintf("Proxy (%s)\n", h.url.String())
}
