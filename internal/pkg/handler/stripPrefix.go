package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
)

type stripPrefix struct {
	next   http.Handler
	prefix string
}

//StripPrefix creates handler
func StripPrefix(next http.Handler, prefix string) http.Handler {
	res := &stripPrefix{}
	res.next = next
	res.prefix = prefix
	return res
}

func (h *stripPrefix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	old := r.URL.Path
	r.URL.Path = strings.TrimPrefix(r.URL.Path, h.prefix)
	if r.URL.Path == old {
		goapp.Log.Warnf("Path '%s' was not stripped by '%s'", old, h.prefix)
	}

	h.next.ServeHTTP(w, r)
}

func (h *stripPrefix) Info(pr string) string {
	return pr + fmt.Sprintf("StripPrefix(%s)\n", h.prefix) + GetInfo(LogShitf(pr), h.next)
}
