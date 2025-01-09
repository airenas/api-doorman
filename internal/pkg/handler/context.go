package handler

import (
	"context"
	"net/http"

	"github.com/airenas/api-doorman/internal/pkg/model"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/oklog/ulid/v2"
)

type customData struct {
	ResponseCode   int
	Key            string
	KeyID          string
	IP             string
	Manual         bool
	QuotaValue     float64
	RateLimitValue int64
	Value          string
	Discount       *bool
	Tags           []string
	RequestID      string
}

func customContext(r *http.Request) (*http.Request, *customData) {
	res, ok := r.Context().Value(model.CtxContext).(*customData)
	if ok {
		return r, res
	}
	res = &customData{}
	res.IP = utils.ExtractIP(r)
	res.RequestID = ulid.Make().String()
	ctx := context.WithValue(r.Context(), model.CtxContext, res)
	return r.WithContext(ctx), res
}
