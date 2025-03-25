package text

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Extractor extract txt from file
type Extractor struct {
	httpclient *http.Client
	timeOut    time.Duration
	url        string
}

// NewExtractor creates a e text extractor instance
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
	res.httpclient = &http.Client{Transport: utils.NewTransport()}
	res.timeOut = time.Minute
	return &res, nil
}

// Get return text by calling the service
func (dc *Extractor) Get(ctx context.Context, name string, file io.Reader) (string, error) {
	ctx, span := utils.StartSpan(ctx, "Extractor.Get", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	if filepath.Ext(name) == ".txt" {
		res, err := io.ReadAll(file)
		if err != nil {
			return "", errors.Wrap(err, "can't read file")
		}
		return string(res), nil
	}

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
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	ctx, cancelF := context.WithTimeout(ctx, dc.timeOut)
	defer cancelF()
	req = req.WithContext(ctx)

	log.Debug().Msgf("Sending file to: %s", dc.url)
	resp, err := dc.httpclient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1000))
		_ = resp.Body.Close()
	}()

	if err := goapp.ValidateHTTPResp(resp, 100); err != nil {
		return "", errors.Wrap(err, "can't get text")
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
