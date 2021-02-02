package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/airenas/go-app/pkg/goapp"
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
	var data map[string]interface{}
	err := json.Unmarshal(bodyBytes, &data)
	if err != nil {
		http.Error(w, "No field "+h.field, http.StatusBadRequest)
		goapp.Log.Error("Can't extract json field. ", err)
		return
	}
	f := data[h.field]
	if f == nil {
		http.Error(w, "No field "+h.field, http.StatusBadRequest)
		goapp.Log.Error("No json field. ")
		return
	}
	var ok bool
	ctx.Value, ok = f.(string)
	if !ok {
		http.Error(w, "Field is not string type "+h.field, http.StatusBadRequest)
		goapp.Log.Errorf("Field is not a string %v", f)
		return
	}
	rn.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	h.next.ServeHTTP(w, rn)
}

func (h *jsonField) Info(pr string) string {
	return pr + fmt.Sprintf("JSONField(%s)\n", h.field) + GetInfo(LogShitf(pr), h.next)
}
