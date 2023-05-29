package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/airenas/go-app/pkg/goapp"
)

type jsonTTSExtract struct {
	next http.Handler
}

// TakeJSONTTS creates handler
func TakeJSONTTS(next http.Handler) http.Handler {
	res := &jsonTTSExtract{}
	res.next = next
	return res
}

type ttsData struct {
	Text             string `json:"text,omitempty"`
	AllowCollectData *bool  `json:"saveRequest,omitempty"`
}

func (h *jsonTTSExtract) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)

	// read all bytes from content body and create new stream using it.
	bodyBytes, _ := io.ReadAll(r.Body)
	var data ttsData
	err := json.Unmarshal(bodyBytes, &data)
	if err != nil {
		http.Error(w, "No field text", http.StatusBadRequest)
		goapp.Log.Error("Can't extract json field. ", err)
		return
	}
	ctx.Value = data.Text
	ctx.Discount = data.AllowCollectData
	rn.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	h.next.ServeHTTP(w, rn)
}

func (h *jsonTTSExtract) Info(pr string) string {
	return pr + fmt.Sprintf("JSONTTSField(%s)\n", "text") + GetInfo(LogShitf(pr), h.next)
}
