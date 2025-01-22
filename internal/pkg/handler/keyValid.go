package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// KeyValidator validator
type KeyValidator interface {
	IsValid(ctx context.Context, key string, ip string, manual bool) (bool /*valid*/, string /*id*/, []string /*tags*/, error)
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
	ok, id, tags, err := h.kv.IsValid(r.Context(), ctx.Key, ctx.IP, ctx.Manual)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Error().Err(err).Msg("can't check key")
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
		log.Error().Err(err).Msg("can't check rate limit setting")
		return
	}
	h.next.ServeHTTP(w, rn)
}

const rateLimitTag = "x-rate-limit:"

func getLimitSetting(tags []string) (int64, error) {
	for _, hs := range tags {
		if strings.HasPrefix(hs, rateLimitTag) {
			str := hs[len(rateLimitTag):]
			return strconv.ParseInt(strings.TrimSpace(str), 10, 64)
		}
	}
	return 0, nil
}

func (h *keyValid) Info(pr string) string {
	return pr + "KeyValid\n" + GetInfo(LogShitf(pr), h.next)
}
