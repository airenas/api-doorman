package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type cleanHeader struct {
	next    http.Handler
	headers []string
}

// CleanHeader removes header with key names starting with 'starting' from request
func CleanHeader(next http.Handler, starting string) (http.Handler, error) {
	res := &cleanHeader{}
	res.next = next
	hdrs := getHeaders(starting)
	if len(hdrs) == 0 {
		return nil, errors.New("no clean headers")
	}
	res.headers = hdrs
	return res, nil
}

func getHeaders(starting string) []string {
	var res []string
	for _, s := range strings.Split(starting, ",") {
		st := strings.TrimSpace(s)
		if st != "" {
			res = append(res, strings.ToUpper(st))
		}
	}
	return res
}

func (h *cleanHeader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, k := range h.headers {
		r.Header.Del(k)
	}
	h.next.ServeHTTP(w, r)
}

func (h *cleanHeader) Info(pr string) string {
	return pr + fmt.Sprintf("CleanHeader (%v)\n", h.headers) + GetInfo(LogShitf(pr), h.next)
}
