package handler

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/airenas/api-doorman/internal/pkg/cmdapp"
)

//AudioLenGetter get duration
type AudioLenGetter interface {
	Get(name string, file io.Reader) (float64, error)
}

type audioLen struct {
	next            http.Handler
	field           string
	durationService AudioLenGetter
}

//AudioLenQuota creates handler
func AudioLenQuota(next http.Handler, field string, srv AudioLenGetter) http.Handler {
	res := &audioLen{}
	res.next = next
	res.field = field
	res.durationService = srv
	return res
}

func (h *audioLen) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	bodyBytes, _ := ioutil.ReadAll(rn.Body)
	rn.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	// create new request for parsing the body
	req2, _ := http.NewRequest(rn.Method, rn.URL.String(), bytes.NewReader(bodyBytes))
	req2.Header = rn.Header
	req2.ParseMultipartForm(32 << 20)
	file, handler, err := req2.FormFile(h.field)
	if err != nil {
		http.Error(w, "No file", http.StatusBadRequest)
		ctx.ResponseCode = http.StatusBadRequest
		cmdapp.Log.Error(err)
		return
	}
	defer file.Close()
	// read all bytes from content body and create new stream using it.
	dur, err := h.durationService.Get(handler.Filename, file)
	if err != nil {
		http.Error(w, "Can't get audio duration", http.StatusInternalServerError)
		ctx.ResponseCode = http.StatusInternalServerError
		cmdapp.Log.Error(err)
		return
	}

	ctx.QuotaValue = dur
	rn.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	if h.next != nil {
		h.next.ServeHTTP(w, rn)
	}
}
