package handler

import (
	"net/http"

	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/sirupsen/logrus"
)

type logStdout struct {
	next  http.Handler
	field string
}

//LogStdout creates handler
func LogStdout(next http.Handler) http.Handler {
	res := &logStdout{}
	res.next = next
	return res
}

func (h *logStdout) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	logrus.Infof("%s %.2f '%s'", utils.ExtractIP(r), ctx.QuotaValue, ctx.Value)
	if h.next != nil {
		h.next.ServeHTTP(w, rn)
	}
}
