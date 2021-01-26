package handler

import (
	"context"
	"net/http"

	"github.com/airenas/api-doorman/internal/pkg/utils"
)

type key int

const (
	// CtxContext context key for custom context object
	CtxContext key = iota
)

type customData struct {
	ResponseCode int
	Key          string
	IP           string
	Manual       bool
	QuotaValue   float64
	Value        string
	Tags         []string
}

func customContext(r *http.Request) (*http.Request, *customData) {
	res, ok := r.Context().Value(CtxContext).(*customData)
	if ok {
		return r, res
	}
	res = &customData{}
	res.IP = utils.ExtractIP(r)
	ctx := context.WithValue(r.Context(), CtxContext, res)
	return r.WithContext(ctx), res
}
