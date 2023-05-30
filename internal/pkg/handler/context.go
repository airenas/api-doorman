package handler

import (
	"context"
	"net/http"

	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/google/uuid"
)

type key int

const (
	// CtxContext context key for custom context object
	CtxContext key = iota
)

type customData struct {
	ResponseCode int
	Key          string
	KeyID        string
	IP           string
	Manual       bool
	QuotaValue   float64
	Value        string
	Discount     *bool
	Tags         []string
	RequestID    string
}

func customContext(r *http.Request) (*http.Request, *customData) {
	res, ok := r.Context().Value(CtxContext).(*customData)
	if ok {
		return r, res
	}
	res = &customData{}
	res.IP = utils.ExtractIP(r)
	res.RequestID = uuid.NewString()
	ctx := context.WithValue(r.Context(), CtxContext, res)
	return r.WithContext(ctx), res
}
