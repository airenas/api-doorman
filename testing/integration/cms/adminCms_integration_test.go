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
	case <-time.After(15 * time.Second):
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
				time.Sleep(time.Second)
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
	checkCode(t, invokeRequest(t, newRequest(t, http.MethodGet, "/live", nil)), http.StatusOK)
}

func TestCreate(t *testing.T) {
	in := api.CreateInput{ID: uuid.NewString(), OperationID: uuid.NewString(), Service: "test", Credits: 100}
	resp := invokeRequest(t, newRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusCreated)
	res := api.Key{}
	decodeResp(t, resp, &res)
	assert.NotEmpty(t, res.Key)

	resp = invokeRequest(t, newRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusConflict)
	resN := api.Key{}
	decodeResp(t, resp, &resN)
	assert.Equal(t, res.Key, resN.Key)

	resp = invokeRequest(t, newRequest(t, http.MethodPost, "/key",
		api.CreateInput{ID: uuid.NewString(), OperationID: in.OperationID,
			Service: "test", Credits: 100}))
	checkCode(t, resp, http.StatusBadRequest)
}

func TestGet(t *testing.T) {
	in := api.CreateInput{ID: uuid.NewString(), OperationID: uuid.NewString(), Service: "test", Credits: 100}
	checkCode(t, invokeRequest(t, newRequest(t, http.MethodPost, "/key", in)), http.StatusCreated)

	resp := invokeRequest(t, newRequest(t, http.MethodGet, "/key/"+in.ID, nil))
	checkCode(t, resp, http.StatusOK)
	res := api.Key{}
	decodeResp(t, resp, &res)
	assert.Equal(t, in.Credits, res.TotalCredits)
	assert.Equal(t, "", res.Key)
}

func TestGet_ReturnKey(t *testing.T) {
	in := api.CreateInput{ID: uuid.NewString(), OperationID: uuid.NewString(), Service: "test", Credits: 100}
	checkCode(t, invokeRequest(t, newRequest(t, http.MethodPost, "/key", in)), http.StatusCreated)

	resp := invokeRequest(t, newRequest(t, http.MethodGet, fmt.Sprintf("/key/%s?returnKey=1", in.ID), nil))
	checkCode(t, resp, http.StatusOK)
	res := api.Key{}
	decodeResp(t, resp, &res)
	assert.NotEmpty(t, res.Key)
}

func TestGet_NotFound(t *testing.T) {
	checkCode(t, invokeRequest(t, newRequest(t, http.MethodGet,
		fmt.Sprintf("/key/%s", uuid.NewString()), nil)), http.StatusBadRequest)
}

func newRequest(t *testing.T, method string, urlSuffix string, body interface{}) *http.Request {
	req, err := http.NewRequest(method, cfg.url+urlSuffix, mocks.ToReader(body))
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if body != nil {
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	return req
}

func invokeRequest(t *testing.T, r *http.Request) *http.Response {
	resp, err := cfg.httpclient.Do(r)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	return resp
}

func checkCode(t *testing.T, resp *http.Response, expected int) {
	if resp.StatusCode != expected {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("status %d != %d, %s", expected, resp.StatusCode, string(b))
	}
}

func decodeResp(t *testing.T, resp *http.Response, to interface{}) {
	if err := json.NewDecoder(resp.Body).Decode(to); err != nil {
		t.Fatalf("json error %v", err)
	}
}
