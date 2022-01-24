package handler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
)

const authHeader = "Authorization"

type keyExtract struct {
	next http.Handler
}

//KeyExtract creates handler
func KeyExtract(next http.Handler) http.Handler {
	res := &keyExtract{}
	res.next = next
	return res
}

//ServeHTTP tries to extract authorization key from header Authorization: Key <key>
// if not there then tries to extract from query parameter key=<key>
func (h *keyExtract) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key, err := extractKey(r.Header.Get(authHeader))
	if err != nil {
		http.Error(w, "Wrong auth header", http.StatusUnauthorized)
		goapp.Log.Error(errors.Wrap(err, "can't extract key from header"))
		return
	}
	rn, ctx := customContext(r)
	if key != "" {
		rn.Header.Del(authHeader)
	} else {
		keys, ok := r.URL.Query()["key"]
		if ok && len(keys[0]) > 0 {
			key = keys[0]
			q, _ := url.ParseQuery(rn.URL.RawQuery)
			q.Del("key")
			rn.URL.RawQuery = q.Encode()
		}
	}
	if key != "" {
		ctx.Key = key
		ctx.Manual = true
	}

	h.next.ServeHTTP(w, rn)
}

func extractKey(str string) (string, error) {
	if str == "" {
		return "", nil
	}
	strs := strings.Split(str, " ")
	if len(strs) != 2 || strs[0] != "Key" {
		return "", errors.New("wrong key format, wanted: Key <key>")
	}
	return strs[1], nil
}

func (h *keyExtract) Info(pr string) string {
	return pr + "KeyExtract\n" + GetInfo(LogShitf(pr), h.next)
}
