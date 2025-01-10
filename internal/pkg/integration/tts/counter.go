package tts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Counter gets usage count for URL
type Counter struct {
	httpclient *http.Client
	timeOut    time.Duration
	url        string
	paramName  string
}

// NewCounter creates a new counter instance
func NewCounter(url string) (*Counter, error) {
	prm, err := extractParam(url)
	if err != nil {
		return nil, err
	}
	res := Counter{}
	res.httpclient = &http.Client{Transport: utils.NewTransport()}
	res.timeOut = time.Second * 10
	res.url = url
	res.paramName = prm
	return &res, nil
}

func extractParam(url string) (string, error) {
	re := regexp.MustCompile("{{.*}}")
	expr := re.FindString(url)
	res := strings.ReplaceAll(strings.ReplaceAll(expr, "{{", ""), "}}", "")
	if res == "" {
		return "", fmt.Errorf("no expr in URL '%s'", url)
	}
	return res, nil
}

// Get return text by calling the service
func (c *Counter) Get(prm string) (int64, error) {
	url := prepareURL(c.url, c.paramName, prm)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	ctx, cancelF := context.WithTimeout(context.Background(), c.timeOut)
	defer cancelF()
	req = req.WithContext(ctx)

	log.Debug().Msgf("Invoke: %s", url)
	resp, err := c.httpclient.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1000))
		_ = resp.Body.Close()
	}()

	if err := goapp.ValidateHTTPResp(resp, 100); err != nil {
		return 0, errors.Wrap(err, "can't get text")
	}
	var respData countResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return 0, errors.Wrap(err, "can't decode response")
	}
	return respData.Count, nil
}

// GetParamName returns parameter name for URL query parameter to use
func (c *Counter) GetParamName() string {
	return c.paramName
}

func prepareURL(s, pName, param string) string {
	return strings.ReplaceAll(s, fmt.Sprintf("{{%s}}", pName), param)
}

type countResponse struct {
	Count int64 `json:"count,omitempty"`
}
