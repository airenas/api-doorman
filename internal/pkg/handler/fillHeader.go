package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const headerSaveTags = "x-tts-save-tags"
const headerRequestID = "x-doorman-requestid"

type fillHeader struct {
	next http.Handler
}

// FillHeader creates handler for filling header values from tags
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
			log.Error().Err(err).Msg("Can't parse header value from tag")
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

// FillKeyHeader creates handler for adding key hash value into "x-tts-save-tags"
func FillKeyHeader(next http.Handler) http.Handler {
	res := &fillKeyHeader{}
	res.next = next
	return res
}

func (h *fillKeyHeader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	if ctx.Key != "" {
		setHeader(rn, headerSaveTags, "key_"+hashKey(ctx.Key))
	}
	h.next.ServeHTTP(w, rn)
}

func (h *fillKeyHeader) Info(pr string) string {
	return pr + "FillKeyHeader\n" + GetInfo(LogShitf(pr), h.next)
}

func hashKey(k string) string {
	h := sha256.New()
	h.Write([]byte(k))
	// trim as we need just to know the request goes from same key
	return trim(hex.EncodeToString(h.Sum(nil)), 10)
}

func trim(s string, i int) string {
	if len(s) > i {
		return s[:i]
	}
	return s
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

type fillRequestIDHeader struct {
	db   string
	next http.Handler
}

// FillRequestIDHeader creates handler for adding requestID into header x-doorman-requestid"
func FillRequestIDHeader(next http.Handler, dbName string) http.Handler {
	res := &fillRequestIDHeader{}
	res.next = next
	res.db = dbName
	return res
}

func (h *fillRequestIDHeader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	if ctx.RequestID != "" {
		setHeader(rn, headerRequestID, fmt.Sprintf("%s:%s:%s", h.db, manualStr(ctx.Manual), ctx.RequestID))
	}
	h.next.ServeHTTP(w, rn)
}

func manualStr(b bool) string {
	if b {
		return "m"
	}
	return ""
}

func (h *fillRequestIDHeader) Info(pr string) string {
	return pr + fmt.Sprintf("FillRequestIDHeader(db:%s)\n", h.db) + GetInfo(LogShitf(pr), h.next)
}
