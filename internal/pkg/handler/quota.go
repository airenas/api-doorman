package handler

import (
	"net/http"
)

type requestAsQuota struct {
	next http.Handler
}

// RequestAsQuota creates handler
func RequestAsQuota(next http.Handler) http.Handler {
	res := &requestAsQuota{}
	res.next = next
	return res
}

func (h *requestAsQuota) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	ctx.QuotaValue = 1
	h.next.ServeHTTP(w, rn)
}

func (h *requestAsQuota) Info(pr string) string {
	return pr + "RequestAsQuota\n" + GetInfo(LogShitf(pr), h.next)
}
