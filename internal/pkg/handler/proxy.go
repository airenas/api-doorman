package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type proxy struct {
	url *url.URL
}

//Proxy creates handler
func Proxy(sURL string) http.Handler {
	res := &proxy{}
	res.url, _ = url.Parse(sURL)
	return res
}

func (h *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(h.url)
	r.URL.Host = h.url.Host
	r.URL.Scheme = h.url.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = h.url.Host
	rn, ctx := customContext(r)
	proxy.ModifyResponse = func(resp *http.Response) (err error) {
		ctx.ResponseCode = resp.StatusCode
		return nil
	}
	proxy.ServeHTTP(w, rn)
}
