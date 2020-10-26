package handler

import (
	"net/http"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/sirupsen/logrus"
)

//DBSaver logs to db
type DBSaver interface {
	Save(*api.Log) error
}

type logDB struct {
	next http.Handler
	dbs  DBSaver
}

//LogDB creates handler
func LogDB(next http.Handler, dbs DBSaver) http.Handler {
	res := &logDB{}
	res.next = next
	res.dbs = dbs
	return res
}

func (h *logDB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	if h.next != nil {
		h.next.ServeHTTP(w, rn)
	}
	data := &api.Log{}
	data.Value = ctx.Value
	data.Date = time.Now()
	data.QuotaValue = ctx.QuotaValue
	data.Key = ctx.Key
	data.IP = utils.ExtractIP(r)
	data.URL = r.URL.String()
	data.ResponseCode = ctx.ResponseCode
	data.Fail = !(data.ResponseCode >= 200 && data.ResponseCode < 300)
	go func() {
		err := h.dbs.Save(data)
		if err != nil {
			logrus.Error("Can't save lod. ", err)
		}
	}()
}
