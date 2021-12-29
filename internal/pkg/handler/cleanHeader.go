package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type cleanHeader struct {
	next     http.Handler
	starting string
}

//FillHeader creates handler for filling header values from tags
func CleanHeader(next http.Handler, starting string) (http.Handler, error) {
	res := &cleanHeader{}
	res.next = next
	st := strings.ToUpper(strings.TrimSpace(starting))
	if st == "" {
		return nil, errors.New("no clean header prefix")
	}
	res.starting = st
	return res, nil
}

func (h *cleanHeader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hs := r.Header
	for k := range hs {
		if strings.HasPrefix(strings.ToUpper(k), h.starting) {
			r.Header.Del(k)
		}
	}
	h.next.ServeHTTP(w, r)
}

func (h *cleanHeader) Info(pr string) string {
	return pr + fmt.Sprintf("CleanHeader (starting:%s)\n", h.starting) + GetInfo(LogShitf(pr), h.next)
}
