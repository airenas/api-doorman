package handler

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// TextGetter get duration
type TextGetter interface {
	Get(name string, file io.Reader) (string, error)
}

type toTextAndQuota struct {
	next           http.Handler
	field          string
	getTextService TextGetter
}

// ToTextAndQuota creates handler. The handler:
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
	bodyBytes, err := io.ReadAll(rn.Body)
	if err != nil {
		http.Error(w, "Can't read request", http.StatusBadRequest)
		log.Error().Err(err).Send()
		return
	}
	rn.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// create new request for parsing the body
	req2, _ := http.NewRequest(rn.Method, rn.URL.String(), bytes.NewReader(bodyBytes))
	req2.Header = rn.Header
	err = req2.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Can't parse form data", http.StatusBadRequest)
		log.Error().Err(err).Send()
		return
	}
	defer cleanFiles(req2.MultipartForm)
	file, handler, err := req2.FormFile(h.field)
	if err != nil {
		http.Error(w, "No file", http.StatusBadRequest)
		ctx.ResponseCode = http.StatusBadRequest
		log.Error().Err(err).Send()
		return
	}
	defer file.Close()

	txt, err := h.getTextService.Get(handler.Filename, file)
	if err != nil {
		http.Error(w, "Can't extract text", http.StatusInternalServerError)
		ctx.ResponseCode = http.StatusInternalServerError
		log.Error().Err(err).Send()
		return
	}

	ctx.QuotaValue = float64(len([]rune(txt)))

	newBytes, hv, err := copyFormData(h.field, handler.Filename, txt, req2.MultipartForm.Value)
	if err != nil {
		http.Error(w, "Can't prepare new body", http.StatusInternalServerError)
		ctx.ResponseCode = http.StatusInternalServerError
		log.Error().Err(err).Send()
		return
	}
	rn.Body = io.NopCloser(newBytes)
	rn.Header.Set("Content-Type", hv)
	rn.Header.Set("Content-Length", strconv.Itoa(newBytes.Len()))
	rn.ContentLength = int64(newBytes.Len())

	h.next.ServeHTTP(w, rn)
}

func copyFormData(field, fileName, str string, formValues map[string][]string) (*bytes.Buffer, string, error) {
	res := &bytes.Buffer{}
	writer := multipart.NewWriter(res)
	defer writer.Close()
	part, err := writer.CreateFormFile(field, toTxtExt(fileName))
	if err != nil {
		return nil, "", errors.Wrapf(err, "can't create form file")
	}
	_, err = io.Copy(part, strings.NewReader(str))
	if err != nil {
		return nil, "", errors.Wrapf(err, "can't write form file")
	}
	for k, vs := range formValues {
		for _, v := range vs {
			err = writer.WriteField(k, v)
			if err != nil {
				return nil, "", errors.Wrapf(err, "can't write form field %s", k)
			}
		}
	}
	return res, writer.FormDataContentType(), nil
}

func toTxtExt(s string) string {
	return strings.TrimSuffix(s, filepath.Ext(s)) + ".txt"
}

func (h *toTextAndQuota) Info(pr string) string {
	return pr + fmt.Sprintf("ToTextAndQuota(%s)\n", h.field) + GetInfo(pr+" ", h.next)
}
