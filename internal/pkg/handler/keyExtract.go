package handler

import (
	"net/http"
	"net/url"
)

type keyExtract struct {
	next http.Handler
}

//KeyExtract creates handler
func KeyExtract(next http.Handler) http.Handler {
	res := &keyExtract{}
	res.next = next
	return res
}

func (h *keyExtract) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["key"]
	rn, ctx := customContext(r)
	if ok && len(keys[0]) > 0 {
		key := keys[0]
		ctx.Key = key
		ctx.Manual = true
		q, _ := url.ParseQuery(rn.URL.RawQuery)
		q.Del("key")
		rn.URL.RawQuery = q.Encode()
	}

	h.next.ServeHTTP(w, rn)
}
