package handler

import (
	"net/http"
)

type jsonAsQuota struct {
	next http.Handler
}

//JSONAsQuota creates handler
func JSONAsQuota(next http.Handler) http.Handler {
	res := &jsonAsQuota{}
	res.next = next
	return res
}

func (h *jsonAsQuota) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	ctx.QuotaValue = float64(len([]rune(ctx.Value)))
	h.next.ServeHTTP(w, rn)
}
