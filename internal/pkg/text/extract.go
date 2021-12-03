package text

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
)

//Extractor extract txt from file
type Extractor struct {
	httpclient *http.Client
	url        string
}

//NewExtractor creates a e text extractor instance
func NewExtractor(urlStr string) (*Extractor, error) {
	res := Extractor{}
	var err error
	urlRes, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "can't parse url "+urlStr)
	}
	if urlRes.Host == "" {
		return nil, errors.New("can't parse url " + urlStr)
	}
	res.url = urlRes.String()
	res.httpclient = &http.Client{}
	return &res, nil
}

//Get return text by calling the service
func (dc *Extractor) Get(name string, file io.Reader) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		return "", errors.Wrap(err, "can't add file to request")
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", errors.Wrap(err, "can't copy file to request")
	}
	writer.Close()
	req, err := http.NewRequest("POST", dc.url, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	ctx, cancelF := context.WithTimeout(context.Background(), time.Minute)
	defer cancelF()
	req = req.WithContext(ctx)

	goapp.Log.Debugf("Sending file to: %s", dc.url)
	resp, err := dc.httpclient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", errors.Errorf("can't invoke %s. Code %d. Response %s", dc.url, resp.StatusCode, string(bodyBytes))
	}
	var respData textResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return "", errors.Wrap(err, "can't decode response")
	}
	return respData.Text, nil
}

type textResponse struct {
	Text string `json:"text"`
}
