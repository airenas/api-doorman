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
	key, _ := r.Context().Value(CtxKey).(string)
	quotaV, _ := r.Context().Value(CtxQuotaValue).(float64)
	logrus.Infof("Url Param 'key' is: " + string(key))
	logrus.Infof("Quota value: %f", quotaV)

	ok, err := h.qv.SaveValidate(key, utils.ExtractIP(r), quotaV)
	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		logrus.Error("Can't save quota/validate key. ", err)
		return
	}
	if !ok {
		http.Error(w, "Quota reached", http.StatusForbidden)
		return
	}

	if h.next != nil {
		h.next.ServeHTTP(w, r)
	}
}
