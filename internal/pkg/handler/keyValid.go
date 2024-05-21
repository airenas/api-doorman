package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
)

// KeyValidator validator
type KeyValidator interface {
	IsValid(string, string, bool) (bool, string, []string, error)
}

type keyValid struct {
	next http.Handler
	kv   KeyValidator
}

// KeyValid creates handler
func KeyValid(next http.Handler, kv KeyValidator) http.Handler {
	res := &keyValid{}
	res.kv = kv
	res.next = next
	return res
}

func (h *keyValid) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	ok, id, tags, err := h.kv.IsValid(ctx.Key, ctx.IP, ctx.Manual)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		goapp.Log.Error("can't check key. ", err)
		return
	}
	if !ok {
		http.Error(w, "Key is not valid", http.StatusUnauthorized)
		return
	}
	ctx.Tags = tags
	ctx.KeyID = id
	if ctx.RateLimitValue, err = getLimitSetting(tags); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		goapp.Log.Error("can't check rate limit setting.", err)
		return
	}
	h.next.ServeHTTP(w, rn)
}

const rateLimitTag = "x-rate-limit:"

func getLimitSetting(tags []string) (int64, error) {
	for _, hs := range tags {
		if strings.HasPrefix(hs, rateLimitTag) {
			str := hs[len(rateLimitTag):]
			return strconv.ParseInt(str, 10, 64)
		}
	}
	return 0, nil
}

func (h *keyValid) Info(pr string) string {
	return pr + "KeyValid\n" + GetInfo(LogShitf(pr), h.next)
}
