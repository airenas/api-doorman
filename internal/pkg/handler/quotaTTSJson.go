package handler

import (
	"fmt"
	"net/http"
)

const (
	allowSaveHeader = "x-tts-collect-data"
	allowSaveValue  = "always"
)

type jsonTTSAsQuota struct {
	next     http.Handler
	discount float64
}

//JSONTTSAsQuota creates handler
func JSONTTSAsQuota(next http.Handler, discount float64) http.Handler {
	res := &jsonTTSAsQuota{}
	res.next = next
	res.discount = discount
	return res
}

func (h *jsonTTSAsQuota) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	ctx.QuotaValue = float64(len([]rune(ctx.Value)))
	d := discount(ctx, h.discount)
	if d < 1 {
		ctx.QuotaValue = ctx.QuotaValue * d
	}
	h.next.ServeHTTP(w, rn)
}

func (h *jsonTTSAsQuota) Info(pr string) string {
	return pr + fmt.Sprintf("jsonTTSAsQuota(discount: %f)\n", h.discount) + GetInfo(LogShitf(pr), h.next)
}

func discount(ctx *customData, def float64) float64 {
	if isDiscount(ctx) {
		return def
	}
	return 1
}

func isDiscount(ctx *customData) bool {
	if ctx.Discount != nil {
		return *ctx.Discount
	}
	for _, t := range ctx.Tags {
		h, v, err := headerValue(t)
		if err == nil && h == allowSaveHeader {
			return v == allowSaveValue
		}
	}
	return false
}
