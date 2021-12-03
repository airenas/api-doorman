package handler

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
)

//TextGetter get duration
type TextGetter interface {
	Get(name string, file io.Reader) (string, error)
}

type toTextAndQuota struct {
	next           http.Handler
	field          string
	getTextService TextGetter
}

//ToTextAndQuota creates handler. The handler:
// - extracts file from form field,
// - converts file to txt,
// - packs text as file into new request
func ToTextAndQuota(next http.Handler, field string, srv TextGetter) http.Handler {
	res := &toTextAndQuota{}
	res.next = next
	res.field = field
	res.getTextService = srv
	return res
}

func (h *toTextAndQuota) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	bodyBytes, _ := ioutil.ReadAll(rn.Body)
	rn.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	// create new request for parsing the body
	req2, _ := http.NewRequest(rn.Method, rn.URL.String(), bytes.NewReader(bodyBytes))
	req2.Header = rn.Header
	err := req2.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Can't parse form data", http.StatusBadRequest)
		goapp.Log.Error(err)
		return
	}
	defer cleanFiles(req2.MultipartForm)
	file, handler, err := req2.FormFile(h.field)
	if err != nil {
		http.Error(w, "No file", http.StatusBadRequest)
		ctx.ResponseCode = http.StatusBadRequest
		goapp.Log.Error(err)
		return
	}
	defer file.Close()
	// read all bytes from content body and create new stream using it.
	txt, err := h.getTextService.Get(handler.Filename, file)
	if err != nil {
		http.Error(w, "Can't extract text", http.StatusInternalServerError)
		ctx.ResponseCode = http.StatusInternalServerError
		goapp.Log.Error(err)
		return
	}

	ctx.QuotaValue = float64(len([]rune(txt)))

	newBytes, err := copyFormData(h.field, handler.Filename, txt, req2.MultipartForm.Value)
	if err != nil {
		http.Error(w, "Can't prepare new body", http.StatusInternalServerError)
		ctx.ResponseCode = http.StatusInternalServerError
		goapp.Log.Error(err)
		return
	}
	rn.Body = ioutil.NopCloser(newBytes)

	h.next.ServeHTTP(w, rn)
}

func copyFormData(field, fileName, str string, formValues map[string][]string) (*bytes.Buffer, error) {
	res := &bytes.Buffer{}
	writer := multipart.NewWriter(res)
	defer writer.Close()
	part, err := writer.CreateFormFile(field, toTxtExt(fileName))
	if err != nil {
		return res, errors.Wrapf(err, "can't create form file")
	}
	_, err = io.Copy(part, strings.NewReader(str))
	if err != nil {
		return res, errors.Wrapf(err, "can't write form file")
	}
	for k, vs := range formValues {
		for _, v := range vs {
			err = writer.WriteField(k, v)
			if err != nil {
				return res, errors.Wrapf(err, "can't write form field %s", k)
			}
		}
	}
	return res, nil
}

func toTxtExt(s string) string {
	return strings.TrimSuffix(s, filepath.Ext(s)) + ".txt"
}

func (h *toTextAndQuota) Info(pr string) string {
	return pr + fmt.Sprintf("ToTextAndQuota(%s)\n", h.field) + GetInfo(pr+" ", h.next)
}
