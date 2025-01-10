package handler

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type fillOutHeader struct {
	next http.Handler
}

// FillOutHeader creates handler for filling header out values from tags
// starting with x-header-out:
func FillOutHeader(next http.Handler) http.Handler {
	res := &fillOutHeader{}
	res.next = next
	return res
}

func (h *fillOutHeader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	for _, hs := range ctx.Tags {
		h, v, err := headerOutValue(hs)
		if err != nil {
			http.Error(w, "Service error", http.StatusInternalServerError)
			log.Error().Err(err).Msg("Can't parse header value from tag")
			return
		}
		if h != "" {
			w.Header().Set(h, v)
		}
	}
	h.next.ServeHTTP(w, rn)
}

const tagStartValue = "x-header-out:"

func headerOutValue(hs string) (string, string, error) {
	if strings.HasPrefix(hs, tagStartValue) {
		kv := strings.TrimSpace(hs[len(tagStartValue):])
		return headerValue(kv)
	}
	return "", "", nil
}

func (h *fillOutHeader) Info(pr string) string {
	return pr + "FillOutHeader\n" + GetInfo(LogShitf(pr), h.next)
}
