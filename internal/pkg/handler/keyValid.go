package handler

import (
	"net/http"

	"github.com/airenas/api-doorman/internal/pkg/cmdapp"
)

//KeyValidator validator
type KeyValidator interface {
	IsValid(string, bool) (bool, error)
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
	ok, err := h.kv.IsValid(ctx.Key, ctx.Manual)
	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		cmdapp.Log.Error("Can't check key. ", err)
		return
	}
	if !ok {
		http.Error(w, "Key is not valid", http.StatusUnauthorized)
		return
	}

	h.next.ServeHTTP(w, rn)
}
