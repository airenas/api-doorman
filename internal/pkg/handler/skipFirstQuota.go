package handler

import (
	"fmt"
	"net/http"

	"github.com/airenas/go-app/pkg/goapp"
)

// CountGetter get usage count from external system
type CountGetter interface {
	Get(id string) (int64, error)
	GetParamName() string
}

type skipFirstQuota struct {
	next    http.Handler
	counter CountGetter
}

//SkipFirstQuota creates handler
func SkipFirstQuota(next http.Handler, cg CountGetter) http.Handler {
	res := &skipFirstQuota{counter: cg}
	res.next = next
	return res
}

func (h *skipFirstQuota) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	pName := h.counter.GetParamName()
	param := r.URL.Query().Get(pName)
	if param == "" {
		http.Error(w, fmt.Sprintf("No param '%s'", pName), http.StatusBadRequest)
		goapp.Log.Error(fmt.Sprintf("no param '%s'", pName))
		return
	}

	count, err := h.counter.Get(param)
	if err != nil {
		http.Error(w, "Can't extract previous count", http.StatusBadRequest)
		goapp.Log.Error(err)
		return
	}
	if count == 0 {
		goapp.Log.Infof("drop quota value - first call")
		ctx.QuotaValue = 0
	}

	h.next.ServeHTTP(w, rn)
}

func (h *skipFirstQuota) Info(pr string) string {
	return pr + fmt.Sprintf("SkipFirstQuota(%s)\n", h.counter.GetParamName()) + GetInfo(LogShitf(pr), h.next)
}
