package handler

import (
	"net/http"
)

type keyValidOrIP struct {
	withKeyHandler http.Handler
	withIPHandler  http.Handler
}

// KeyValidOrIP creates handler
func KeyValidOrIP(withKeyHandler http.Handler, withIPHandler http.Handler) http.Handler {
	res := &keyValidOrIP{}
	res.withKeyHandler = withKeyHandler
	res.withIPHandler = withIPHandler
	return res
}

func (h *keyValidOrIP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	if ctx.Key == "" {
		h.withIPHandler.ServeHTTP(w, rn)
	} else {
		h.withKeyHandler.ServeHTTP(w, rn)
	}
}

func (h *keyValidOrIP) Info(pr string) string {
	return pr + "KeyValidOrIP\n" +
		LogShitf(pr) + "Key:\n" + GetInfo(LogShitf(LogShitf(pr)), h.withKeyHandler) +
		LogShitf(pr) + "IP:\n" + GetInfo(LogShitf(LogShitf(pr)), h.withIPHandler)
}
