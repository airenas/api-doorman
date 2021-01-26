package handler

import (
	"net/http"

	"github.com/airenas/go-app/pkg/goapp"
)

//KeyValidator validator
type KeyValidator interface {
	IsValid(string, string, bool) (bool, []string, error)
}

type keyValid struct {
	next http.Handler
	kv   KeyValidator
}

//KeyValid creates handler
func KeyValid(next http.Handler, kv KeyValidator) http.Handler {
	res := &keyValid{}
	res.kv = kv
	res.next = next
	return res
}

func (h *keyValid) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	ok, tags, err := h.kv.IsValid(ctx.Key, ctx.IP, ctx.Manual)
	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		goapp.Log.Error("Can't check key. ", err)
		return
	}
	if !ok {
		http.Error(w, "Key is not valid", http.StatusUnauthorized)
		return
	}
	ctx.Tags = tags

	h.next.ServeHTTP(w, rn)
}

func (h *keyValid) Info(pr string) string {
	return pr + "KeyValid\n" + GetInfo(LogShitf(pr), h.next)
}
