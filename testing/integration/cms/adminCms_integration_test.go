//go:build integration
// +build integration

package cms

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type config struct {
	url        string
	httpclient *http.Client
}

var cfg config

func TestMain(m *testing.M) {
	cfg.url = os.Getenv("ADMIN_URL")
	if cfg.url == "" {
		log.Fatal("no ADMIN_URL set")
	}
	cfg.httpclient = &http.Client{Timeout: time.Second}
	select {
	case <-time.After(20 * time.Second):
		log.Fatalf("can't access %s", cfg.url)
	case <-waitForReady(cfg.url):
	}
	os.Exit(m.Run())
}

func waitForReady(url string) <-chan struct{} {
	res := make(chan struct{}, 1)
	go func() {
		for {
			if err := listens(url); err != nil {
				log.Printf("waiting for %s ...", url)
				time.Sleep(500 * time.Millisecond)
			} else {
				res <- struct{}{}
				return
			}
		}
	}()
	return res
}

func listens(url string) error {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	_, err := cfg.httpclient.Do(req)
	return err
}

func TestLive(t *testing.T) {
	t.Parallel()
	checkCode(t, invoke(t, newRequest(t, http.MethodGet, "/live", nil)), http.StatusOK)
}

func TestCreate(t *testing.T) {
	t.Parallel()
	in := api.CreateInput{ID: uuid.NewString(), OperationID: uuid.NewString(), Service: "test", Credits: 100}
	resp := invoke(t, newRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusCreated)
	res := api.Key{}
	decode(t, resp, &res)
	assert.NotEmpty(t, res.Key)

	resp = invoke(t, newRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusConflict)
	resN := api.Key{}
	decode(t, resp, &resN)
	assert.Equal(t, res.Key, resN.Key)

	resp = invoke(t, newRequest(t, http.MethodPost, "/key",
		api.CreateInput{ID: uuid.NewString(), OperationID: in.OperationID,
			Service: "test", Credits: 100}))
	checkCode(t, resp, http.StatusBadRequest)
}

func TestGet(t *testing.T) {
	t.Parallel()
	in := api.CreateInput{ID: uuid.NewString(), OperationID: uuid.NewString(), Service: "test", Credits: 100}
	checkCode(t, invoke(t, newRequest(t, http.MethodPost, "/key", in)), http.StatusCreated)

	resp := invoke(t, newRequest(t, http.MethodGet, "/key/"+in.ID, nil))
	checkCode(t, resp, http.StatusOK)
	res := api.Key{}
	decode(t, resp, &res)
	assert.Equal(t, in.Credits, res.TotalCredits)
	assert.Equal(t, "", res.Key)
}

func TestGet_ReturnsKey(t *testing.T) {
	t.Parallel()
	in := api.CreateInput{ID: uuid.NewString(), OperationID: uuid.NewString(), Service: "test", Credits: 100}
	checkCode(t, invoke(t, newRequest(t, http.MethodPost, "/key", in)), http.StatusCreated)

	resp := invoke(t, newRequest(t, http.MethodGet, fmt.Sprintf("/key/%s?returnKey=1", in.ID), nil))
	checkCode(t, resp, http.StatusOK)
	res := api.Key{}
	decode(t, resp, &res)
	assert.NotEmpty(t, res.Key)
}

func TestKey_NotFound(t *testing.T) {
	t.Parallel()
	checkCode(t, invoke(t, newRequest(t, http.MethodGet,
		fmt.Sprintf("/key/%s", uuid.NewString()), nil)), http.StatusBadRequest)
}

func newRequest(t *testing.T, method string, urlSuffix string, body interface{}) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, cfg.url+urlSuffix, mocks.ToReader(body))
	require.Nil(t, err, "not nil error = %v", err)
	if body != nil {
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	return req
}

func invoke(t *testing.T, r *http.Request) *http.Response {
	t.Helper()
	resp, err := cfg.httpclient.Do(r)
	require.Nil(t, err, "not nil error = %v", err)
	return resp
}

func checkCode(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		b, _ := ioutil.ReadAll(resp.Body)
		require.Equal(t, expected, resp.StatusCode, string(b))
	}
}

func decode(t *testing.T, resp *http.Response, to interface{}) {
	t.Helper()
	require.Nil(t, json.NewDecoder(resp.Body).Decode(to))
}
