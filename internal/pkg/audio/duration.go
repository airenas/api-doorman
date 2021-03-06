package audio

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
)

//Duration comunicates with duration service
type Duration struct {
	httpclient *http.Client
	url        string
}

//NewDurationClient creates a transcriber client
func NewDurationClient(urlStr string) (*Duration, error) {
	res := Duration{}
	var err error
	urlRes, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't parse url "+urlStr)
	}
	if urlRes.Host == "" {
		return nil, errors.New("Can't parse url " + urlStr)
	}
	res.url = urlRes.String()
	res.httpclient = &http.Client{}
	return &res, nil
}

//Get return duration by calling the service
func (dc *Duration) Get(name string, file io.Reader) (float64, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		return 0, errors.Wrap(err, "Can't add file to request")
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return 0, errors.Wrap(err, "Can't add file to request")
	}
	writer.Close()
	req, err := http.NewRequest("POST", dc.url, body)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	goapp.Log.Debugf("Sending audio to: %s", dc.url)
	resp, err := dc.httpclient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return 0, errors.New("Can't get duration")
	}
	var respData durationResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return 0, errors.Wrap(err, "Can't decode response")
	}
	return respData.Duration, nil
}

type durationResponse struct {
	Duration float64 `json:"duration"`
}
