package handler

import (
	"net/http"
	"unicode/utf8"
)

type jsonAsQuota struct {
	next http.Handler
}

// JSONAsQuota creates handler
func JSONAsQuota(next http.Handler) http.Handler {
	res := &jsonAsQuota{}
	res.next = next
	return res
}

func (h *jsonAsQuota) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	ctx.QuotaValue = float64(utf8.RuneCountInString(ctx.Value))
	h.next.ServeHTTP(w, rn)
}

func (h *jsonAsQuota) Info(pr string) string {
	return pr + "JSONAsQuota\n" + GetInfo(LogShitf(pr), h.next)
}
