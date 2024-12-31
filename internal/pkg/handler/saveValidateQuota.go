package handler

import (
	"context"
	"fmt"
	"math"
	"net/http"

	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/rs/zerolog/log"
)

// QuotaValidator validator
type QuotaValidator interface {
	SaveValidate(ctx context.Context, key string, ip string, manual bool, quota float64) (bool /*ok*/, float64 /*remainding*/, float64 /*total*/, error)
	Restore(ctx context.Context, key string, manual bool, quota float64) (float64 /*remainding*/, float64 /*total*/, error)
}

type quotaSaveValidate struct {
	next http.Handler
	qv   QuotaValidator
}

// QuotaValidate creates handler
func QuotaValidate(next http.Handler, qv QuotaValidator) http.Handler {
	res := &quotaSaveValidate{}
	res.qv = qv
	res.next = next
	return res
}

func (h *quotaSaveValidate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	quotaV := ctx.QuotaValue
	log.Ctx(rn.Context()).Debug().Float64("value", quotaV).Msg("Using quota")

	ok, rem, tot, err := h.qv.SaveValidate(rn.Context(), ctx.Key, utils.ExtractIP(rn), ctx.Manual, quotaV)
	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		log.Ctx(rn.Context()).Error().Err(err).Msg("Can't save quota/validate key")
		ctx.ResponseCode = http.StatusInternalServerError
		return
	}
	if rem >= 0 {
		w.Header().Set("X-Rate-Limit-Remaining", fmt.Sprintf("%.0f", rem))
	}
	if tot >= 0 {
		w.Header().Set("X-Rate-Limit-Limit", fmt.Sprintf("%.0f", tot))
	}
	if !ok {
		http.Error(w, "Quota reached", http.StatusForbidden)
		ctx.ResponseCode = http.StatusForbidden
		return
	}
	h.next.ServeHTTP(w, rn)

	if isServiceFailure(ctx.ResponseCode) {
		h.tryRestoreQuota(w, rn, ctx)
	}
}

func (h *quotaSaveValidate) Info(pr string) string {
	return pr + "QuotaSaveValidate\n" + GetInfo(LogShitf(pr), h.next)
}

func isServiceFailure(code int) bool {
	return code >= 400 && code < 600
}

func (h *quotaSaveValidate) tryRestoreQuota(w http.ResponseWriter, req *http.Request, ctx *customData) {
	quotaV := ctx.QuotaValue
	log.Ctx(req.Context()).Debug().Float64("value", quotaV).Msg("Try restore quota")

	rem, tot, err := h.qv.Restore(req.Context(), ctx.Key, ctx.Manual, quotaV)
	if err != nil {
		log.Ctx(req.Context()).Error().Err(err).Msg("Can't restore quota")
		return
	}
	w.Header().Set("X-Rate-Limit-Remaining", fmt.Sprintf("%.0f", math.Max(0, rem)))
	w.Header().Set("X-Rate-Limit-Limit", fmt.Sprintf("%.0f", math.Max(0, tot)))
}
