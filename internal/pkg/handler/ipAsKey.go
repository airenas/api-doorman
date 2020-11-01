package handler

import (
	"net/http"

	"github.com/airenas/api-doorman/internal/pkg/cmdapp"
	"github.com/airenas/api-doorman/internal/pkg/utils"
)

//IPSaver saves ip a=as key into DB
type IPSaver interface {
	Save(string) error
}

type ipAsKey struct {
	next    http.Handler
	ipSaver IPSaver
}

//IPAsKey creates handler
func IPAsKey(next http.Handler, ipSaver IPSaver) http.Handler {
	res := &ipAsKey{}
	res.next = next
	res.ipSaver = ipSaver
	return res
}

func (h *ipAsKey) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	key := utils.ExtractIP(r)
	ctx.Key = key
	err := h.ipSaver.Save(key)
	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		cmdapp.Log.Error("Can't save ip as key. ", err)
		return
	}
	h.next.ServeHTTP(w, rn)
}
