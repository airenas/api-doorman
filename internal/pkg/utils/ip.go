package utils

import (
	"net/http"
	"strings"
)

//ExtractIP return ip from request
func ExtractIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return trimPort(forwarded)
	}
	return trimPort(r.RemoteAddr)
}

func trimPort(s string) string {
	return strings.Split(s, ":")[0]
}
