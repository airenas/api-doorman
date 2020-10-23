package handler

import (
	"context"
	"net/http"
)

type requestAsQuota struct {
	next http.Handler
}

//RequestAsQuota creates handler
func RequestAsQuota(next http.Handler) http.Handler {
	res := &requestAsQuota{}
	res.next = next
	return res
}

func (h *requestAsQuota) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, CtxQuotaValue, float64(1))
	if h.next != nil {
		h.next.ServeHTTP(w, r.WithContext(ctx))
	}
}
