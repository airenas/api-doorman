package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/airenas/go-app/pkg/goapp"
)

// AudioLenGetter get duration
type AudioLenGetter interface {
	Get(name string, file io.Reader) (float64, error)
}

type audioLen struct {
	next            http.Handler
	field           string
	durationService AudioLenGetter
}

// AudioLenQuota creates handler
func AudioLenQuota(next http.Handler, field string, srv AudioLenGetter) http.Handler {
	res := &audioLen{}
	res.next = next
	res.field = field
	res.durationService = srv
	return res
}

func (h *audioLen) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn, ctx := customContext(r)
	tmpFileName, closeF, err := saveTempData(rn.Body)
	if err != nil {
		ctx.ResponseCode = writeBadRequestOrInternalError(w, "")
		goapp.Log.Error(err)
		return
	}
	defer closeF()

	dur, badReqMsg, err := h.getDuration(rn, tmpFileName)
	if err != nil {
		ctx.ResponseCode = writeBadRequestOrInternalError(w, badReqMsg)
		goapp.Log.Error(err)
		return
	}
	ctx.QuotaValue = dur

	tmpFile, err := os.Open(tmpFileName)
	if err != nil {
		ctx.ResponseCode = writeBadRequestOrInternalError(w, "")
		goapp.Log.Error(err)
		return
	}
	defer tmpFile.Close()

	rn.Body = io.NopCloser(tmpFile)

	h.next.ServeHTTP(w, rn)
}

func writeBadRequestOrInternalError(w http.ResponseWriter, badReqMsg string) int {
	res, msg := http.StatusBadRequest, badReqMsg
	if msg == "" {
		res, msg = http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)
	}
	http.Error(w, msg, res)
	return res
}

func (h *audioLen) getDuration(rn *http.Request, tmpFileName string) (float64, string, error) {
	tmpFile, err := os.Open(tmpFileName)
	if err != nil {
		return 0, "", fmt.Errorf("can't read file: %w", err)
	}
	defer tmpFile.Close()

	// create new request for parsing the body
	req2, err := http.NewRequest(rn.Method, rn.URL.String(), tmpFile)
	if err != nil {
		return 0, "", fmt.Errorf("can't create request: %w", err)
	}
	req2.Header = rn.Header
	err = req2.ParseMultipartForm(32 << 20)
	if err != nil {
		return 0, "Can't parse form data", fmt.Errorf("can't parse form data: %w", err)
	}
	defer cleanFiles(req2.MultipartForm)
	file, handler, err := req2.FormFile(h.field)
	if err != nil {
		return 0, "No file", fmt.Errorf("no file: %w", err)
	}
	defer file.Close()
	dur, err := h.durationService.Get(handler.Filename, file)
	if err != nil {
		return 0, "", fmt.Errorf("can't get duration: %w", err)
	}
	return dur, "", nil
}

func saveTempData(reader io.Reader) (string, func(), error) {
	tempFile, err := os.CreateTemp("", "doorman/input-body*")
	if err != nil {
		return "", func() {}, err
	}
	defer tempFile.Close()

	delF := func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			goapp.Log.Error(err)
		}
	}
	if _, err := io.Copy(tempFile, reader); err != nil {
		delF()
		return "", func() {}, fmt.Errorf("can't write: %w", err)
	}
	return tempFile.Name(), delF, nil
}

func cleanFiles(f *multipart.Form) {
	if f != nil {
		if err := f.RemoveAll(); err != nil {
			goapp.Log.Error(err)
		}
	}
}

func (h *audioLen) Info(pr string) string {
	return pr + fmt.Sprintf("AudioLenQuota(%s)\n", h.field) + GetInfo(pr+" ", h.next)
}
