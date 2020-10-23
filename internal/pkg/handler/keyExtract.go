package handler

import (
	"context"
	"net/http"
)

//KeyExtract extract key from request and put into context
type KeyExtract struct {
	next http.Handler
}

//NewKeyExtract creates handler
func NewKeyExtract(next http.Handler) *KeyExtract {
	res := &KeyExtract{}
	res.next = next
	return res
}

func (h *KeyExtract) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["key"]

	ctx := r.Context()
	if ok && len(keys[0]) > 0 {
		key := keys[0]
		ctx = context.WithValue(ctx, CtxKey, key)
	}
	if h.next != nil {
		h.next.ServeHTTP(w, r.WithContext(ctx))
	}
}
