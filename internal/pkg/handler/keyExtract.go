package handler

import (
	"net/http"
)

//KeyExtract extract key from request and put into context
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
	}
	if h.next != nil {
		h.next.ServeHTTP(w, rn)
	}
}
