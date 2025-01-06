package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/rs/zerolog/log"
)

// DBSaver logs to db
type DBSaver interface {
	SaveLog(ctx context.Context, data *api.Log) error
}

type logDB struct {
	next http.Handler
	dbs  DBSaver
	sync bool
}

// LogDB creates handler
func LogDB(next http.Handler, dbs DBSaver, syncLog bool) http.Handler {
	res := &logDB{}
	res.next = next
	res.dbs = dbs
	res.sync = syncLog
	return res
}

func (h *logDB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	if h.next != nil {
		h.next.ServeHTTP(w, rn)
	}
	data := &api.Log{}
	// data.Value = ctx.Value
	data.Date = time.Now()
	data.QuotaValue = ctx.QuotaValue
	data.KeyID = strings.TrimSpace(ctx.KeyID)
	data.RequestID = ctx.RequestID
	data.IP = utils.ExtractIP(r)
	data.URL = rn.URL.String()
	data.ResponseCode = ctx.ResponseCode
	data.Fail = responseCodeIsFail(data.ResponseCode)
	sf := func() {
		ctx, cf := context.WithTimeout(context.Background(), 5*time.Second) // use another context, request context can be canceled
		defer cf()
		err := h.dbs.SaveLog(ctx, data)
		if err != nil {
			log.Error().Err(err).Msg("can't save log")
		}
	}
	if h.sync {
		sf()
	} else {
		go func() { sf() }()
	}
}

func responseCodeIsFail(code int) bool {
	return !(code >= 200 && code < 300)
}

func (h *logDB) Info(pr string) string {
	return GetInfo(LogShitf(pr), h.next) + pr + "LogDB\n"
}
