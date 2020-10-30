package handler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

type jsonField struct {
	next  http.Handler
	field string
}

//TakeJSON creates handler
func TakeJSON(next http.Handler, field string) http.Handler {
	res := &jsonField{}
	res.next = next
	res.field = field
	return res
}

func (h *jsonField) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)

	// read all bytes from content body and create new stream using it.
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	var data map[string]string
	err := json.Unmarshal(bodyBytes, &data)
	if err != nil {
		http.Error(w, "No field "+h.field, http.StatusBadRequest)
		logrus.Error("Can't extract json field. ", err)
		return
	}
	f := data[h.field]
	ctx.Value = f
	rn.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	h.next.ServeHTTP(w, rn)
}
