package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/airenas/api-doorman/internal/pkg/utils"
)

type logStdout struct {
	next http.Handler
	log  io.Writer
}

// LogStdout creates handler
func LogStdout(next http.Handler) http.Handler {
	res := &logStdout{}
	res.next = next
	res.log = os.Stdout
	return res
}

func (h *logStdout) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	fmt.Fprintf(h.log, "%s %.2f '%s'", utils.ExtractIP(r), ctx.QuotaValue, ctx.Value)
	h.next.ServeHTTP(w, rn)
}

func (h *logStdout) Info(pr string) string {
	return pr + "LogStdout\n" + GetInfo(LogShitf(pr), h.next)
}
