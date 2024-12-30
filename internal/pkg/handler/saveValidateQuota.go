package handler

import (
	"fmt"
	"math"
	"net/http"

	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/rs/zerolog/log"
)

// QuotaValidator validator
type QuotaValidator interface {
	SaveValidate(string, string, bool, float64) (bool, float64, float64, error)
	Restore(string, bool, float64) (float64, float64, error)
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
	log.Debug().Msgf("Quota value: %f", quotaV)

	ok, rem, tot, err := h.qv.SaveValidate(ctx.Key, utils.ExtractIP(rn), ctx.Manual, quotaV)
	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Can't save quota/validate key")
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

func (h *quotaSaveValidate) tryRestoreQuota(w http.ResponseWriter, rn *http.Request, ctx *customData) {
	quotaV := ctx.QuotaValue
	log.Debug().Msgf("Try restore quota value: %f", quotaV)

	rem, tot, err := h.qv.Restore(ctx.Key, ctx.Manual, quotaV)
	if err != nil {
		log.Error().Err(err).Msg("Can't restore quota")
		return
	}
	w.Header().Set("X-Rate-Limit-Remaining", fmt.Sprintf("%.0f", math.Max(0, rem)))
	w.Header().Set("X-Rate-Limit-Limit", fmt.Sprintf("%.0f", math.Max(0, tot)))
}
