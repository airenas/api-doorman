package utils

import (
	"net/http"
	"strings"
)

//ExtractIP return ip from request
func ExtractIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return strings.Split(forwarded, ":")[0]
	}
	return r.RemoteAddr
}
