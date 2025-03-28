package handler

import (
	"context"
	"net/http"

	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/rs/zerolog/log"
)

// IPSaver saves ip a=as key into DB
type IPSaver interface {
	Save(ctx context.Context, ip string) (string, error)
}

type ipAsKey struct {
	next    http.Handler
	ipSaver IPSaver
}

// IPAsKey creates handler
func IPAsKey(next http.Handler, ipSaver IPSaver) http.Handler {
	res := &ipAsKey{}
	res.next = next
	res.ipSaver = ipSaver
	return res
}

func (h *ipAsKey) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	key := utils.ExtractIP(r)
	log.Debug().Msgf("IP: %s, IP header: '%s'", key, utils.GetIPHeader(r))
	ctx.Key = key
	id, err := h.ipSaver.Save(rn.Context(), key)
	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		log.Error().Err(err).Msg("can't save ip as key")
		return
	}
	ctx.KeyID = id
	h.next.ServeHTTP(w, rn)
}

func (h *ipAsKey) Info(pr string) string {
	return pr + "IPAsKey\n" + GetInfo(LogShitf(pr), h.next)
}
