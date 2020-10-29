package handler

import (
	"context"
	"net/http"
)

type key int

const (
	// CtxContext context key for custom context object
	CtxContext key = iota
)

type customData struct {
	ResponseCode int
	Key          string
	Manual       bool
	QuotaValue   float64
	Value        string
}

func customContext(r *http.Request) (*http.Request, *customData) {
	res, ok := r.Context().Value(CtxContext).(*customData)
	if ok {
		return r, res
	}
	res = &customData{}
	ctx := context.WithValue(r.Context(), CtxContext, res)
	return r.WithContext(ctx), res
}
