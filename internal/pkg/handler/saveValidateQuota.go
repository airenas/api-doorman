package handler

import (
	"net/http"

	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/sirupsen/logrus"
)

//QuotaValidator validator
type QuotaValidator interface {
	SaveValidate(string, string, float64) (bool, error)
}

type quotaSaveValidate struct {
	next http.Handler
	qv   QuotaValidator
}

//QuotaValidate creates handler
func QuotaValidate(next http.Handler, qv QuotaValidator) http.Handler {
	res := &quotaSaveValidate{}
	res.qv = qv
	res.next = next
	return res
}

func (h *quotaSaveValidate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	quotaV := ctx.QuotaValue
	logrus.Infof("Quota value: %f", quotaV)

	ok, err := h.qv.SaveValidate(ctx.Key, utils.ExtractIP(rn), quotaV)
	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		logrus.Error("Can't save quota/validate key. ", err)
		ctx.ResponseCode = http.StatusInternalServerError
		return
	}
	if !ok {
		http.Error(w, "Quota reached", http.StatusForbidden)
		ctx.ResponseCode = http.StatusForbidden
		return
	}

	if h.next != nil {
		h.next.ServeHTTP(w, rn)
	}
}
