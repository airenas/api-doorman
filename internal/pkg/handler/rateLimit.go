package handler

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

// RateLimit validator
type RateLimitValidator interface {
	Validate(string, int64, int64) (bool, int64, int64, error)
}

type rateLimitValidate struct {
	next  http.Handler
	qv    RateLimitValidator
	limit int64
}

// QuotaValidate creates handler
func RateLimitValidate(next http.Handler, qv RateLimitValidator, limit int64) http.Handler {
	res := &rateLimitValidate{}
	res.qv = qv
	res.next = next
	res.limit = limit
	return res
}

func (h *rateLimitValidate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	quotaV := ctx.QuotaValue
	limit := int64(ctx.RateLimitValue)
	if limit <= 0 {
		limit = h.limit
	}
	ok, rem, retryAfter, err := h.qv.Validate(makeRateLimitKey(idOrHash(ctx), ctx.Manual), int64(limit), int64(quotaV))
	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		log.Error().Msgf("can't validate rate limit.", err)
		ctx.ResponseCode = http.StatusInternalServerError
		return
	}
	log.Debug().Msgf("Quota value: %.2f, rem: %d, time: %d, rate limit: %d", quotaV, rem, retryAfter, limit)
	if rem >= 0 {
		w.Header().Set("X-Rate-Limit-Short-Remaining", fmt.Sprintf("%d", rem))
	}
	if retryAfter > 0 {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
	}
	if !ok {
		http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		ctx.ResponseCode = http.StatusTooManyRequests
		return
	}
	h.next.ServeHTTP(w, rn)
}

func idOrHash(ctx *customData) string {
	if ctx.KeyID != "" {
		return ctx.KeyID
	}
	return hashKey(ctx.Key)
}

func makeRateLimitKey(key string, manual bool) string {
	return fmt.Sprintf("%s:%t", key, manual)
}

func (h *rateLimitValidate) Info(pr string) string {
	rStr := "no limiter"
	if ip, ok := h.qv.(infoProvider); ok {
		rStr = ip.Info("")
	}
	return pr + fmt.Sprintf("RateLimitValidate(%d, %s)\n", h.limit, rStr) + GetInfo(LogShitf(pr), h.next)
}
