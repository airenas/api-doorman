package handler

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
)

const headerSaveTags = "x-tts-save-tags"

type fillHeader struct {
	next http.Handler
}

//FillHeader creates handler for filling header values from tags
func FillHeader(next http.Handler) http.Handler {
	res := &fillHeader{}
	res.next = next
	return res
}

func (h *fillHeader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	for _, hs := range ctx.Tags {
		h, v, err := headerValue(hs)
		if err != nil {
			http.Error(w, "Service error", http.StatusInternalServerError)
			goapp.Log.Error("Can't parse header value from tag", err)
			return
		}
		if h != "" {
			rn.Header.Set(h, v)
		}
	}
	h.next.ServeHTTP(w, rn)
}

func headerValue(hs string) (string, string, error) {
	if idx := strings.IndexByte(hs, ':'); idx >= 0 {
		return strings.TrimSpace(hs[:idx]), strings.TrimSpace(hs[idx+1:]), nil
	}
	if (strings.TrimSpace(hs)) == "" {
		return "", "", nil
	}
	return "", "", errors.Errorf("Wrong header value, no ':' in '%s'", hs)
}

func (h *fillHeader) Info(pr string) string {
	return pr + "FillHeader\n" + GetInfo(LogShitf(pr), h.next)
}

type fillKeyHeader struct {
	next http.Handler
}

//FillHeader creates handler for filling header values from tags
func FillKeyHeader(next http.Handler) http.Handler {
	res := &fillKeyHeader{}
	res.next = next
	return res
}

func (h *fillKeyHeader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	if ctx.Key != "" {
		setHeader(rn, headerSaveTags, "key_" + hashKey(ctx.Key))
	}
	h.next.ServeHTTP(w, rn)
}

func (h *fillKeyHeader) Info(pr string) string {
	return pr + "FillKeyHeader\n" + GetInfo(LogShitf(pr), h.next)
}

func hashKey(k string) string {
	h := md5.New()
	h.Write([]byte(k))
	return hex.EncodeToString(h.Sum(nil))
}

func setHeader(r *http.Request, k, v string) {
	if v == "" {
		return
	}
	h := r.Header
	vo := h.Get(k)
	vn := v
	if strings.TrimSpace(vo) != "" {
		vn = vo + "," + v
	}
	h.Set(k, vn)
}
