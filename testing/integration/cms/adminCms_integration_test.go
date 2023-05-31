//go:build integration
// +build integration

package cms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
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
	httpClient *http.Client
}

var cfg config

func TestMain(m *testing.M) {
	cfg.url = os.Getenv("ADMIN_URL")
	if cfg.url == "" {
		log.Fatal("FAIL: no ADMIN_URL set")
	}
	cfg.httpClient = &http.Client{Timeout: time.Second}

	tCtx, cf := context.WithTimeout(context.Background(), time.Second*20)
	defer cf()
	waitForOpenOrFail(tCtx, cfg.url)

	os.Exit(m.Run())
}

func waitForOpenOrFail(ctx context.Context, urlWait string) {
	u, err := url.Parse(urlWait)
	if err != nil {
		log.Fatalf("FAIL: can't parse %s", urlWait)
	}
	for {
		if err := listen(net.JoinHostPort(u.Hostname(), u.Port())); err != nil {
			log.Printf("waiting for %s ...", urlWait)
		} else {
			return
		}
		select {
		case <-ctx.Done():
			log.Fatalf("FAIL: can't access %s", urlWait)
			break
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func listen(urlStr string) error {
	log.Printf("dial %s", urlStr)
	conn, err := net.DialTimeout("tcp", urlStr, time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
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
	assert.Equal(t, "", resN.Key)

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

func TestGet_ReturnsKeyID(t *testing.T) {
	t.Parallel()
	in := api.CreateInput{ID: uuid.NewString(), OperationID: uuid.NewString(), Service: "test", Credits: 100}
	resp := invoke(t, newRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusCreated)
	res := api.Key{}
	decode(t, resp, &res)

	resp = invoke(t, newRequest(t, http.MethodPost, "/keyID", api.Key{Key: res.Key}))
	checkCode(t, resp, http.StatusOK)
	resKey := api.KeyID{}
	decode(t, resp, &resKey)

	assert.Equal(t, in.ID, resKey.ID)
	assert.Equal(t, in.Service, resKey.Service)
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

func TestKeysChanges(t *testing.T) {
	t.Parallel()

	from := time.Now().Add(-time.Millisecond) // make sure we are in past at least by 1ms
	id := uuid.NewString()
	in := api.CreateInput{ID: id, OperationID: uuid.NewString(), Service: "changes", Credits: 100}
	resp := invoke(t, newRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusCreated)

	resp = invoke(t, newRequest(t, http.MethodGet, "/keys/changes", in))
	checkCode(t, resp, http.StatusOK)
	res := api.Changes{}
	decode(t, resp, &res)
	require.NotEmpty(t, filter(res.Data, "changes"))
	assert.Equal(t, id, filter(res.Data, "changes")[0].ID, "%v", res)
	assert.Nil(t, res.From)
	assert.NotNil(t, res.Till)

	in = api.CreateInput{ID: uuid.NewString(), OperationID: uuid.NewString(), Service: "changes", Credits: 100}
	resp = invoke(t, newRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusCreated)
	resp = invoke(t, newRequest(t, http.MethodGet, "/keys/changes", in))
	checkCode(t, resp, http.StatusOK)

	res = api.Changes{}
	decode(t, resp, &res)
	assert.Equal(t, 2, len(filter(res.Data, "changes")))

	resp = invoke(t, newRequest(t, http.MethodGet, fmt.Sprintf("/keys/changes?from=%s", from.Format(time.RFC3339Nano)), in))
	checkCode(t, resp, http.StatusOK)
	res = api.Changes{}
	decode(t, resp, &res)
	assert.Equal(t, 2, len(filter(res.Data, "changes")))
	assert.Equal(t, from.Unix(), res.From.Unix())

	resp = invoke(t, newRequest(t, http.MethodGet, fmt.Sprintf("/keys/changes?from=%s",
		res.Till.Add(time.Millisecond).Format(time.RFC3339Nano)), in))
	checkCode(t, resp, http.StatusOK)
	res = api.Changes{}
	decode(t, resp, &res)
	assert.Equal(t, 0, len(filter(res.Data, "changes")))
	// create one more and see if it appears in changes list
	in = api.CreateInput{ID: uuid.NewString(), OperationID: uuid.NewString(), Service: "changes", Credits: 100}
	resp = invoke(t, newRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusCreated)
	resp = invoke(t, newRequest(t, http.MethodGet, fmt.Sprintf("/keys/changes?from=%s",
		res.Till.Add(time.Millisecond).Format(time.RFC3339Nano)), in))
	checkCode(t, resp, http.StatusOK)
	res = api.Changes{}
	decode(t, resp, &res)
	assert.Equal(t, 1, len(filter(res.Data, "changes")))
}

func filter(keys []*api.Key, s string) []*api.Key {
	var res []*api.Key
	for _, d := range keys {
		if d.Service == "changes" {
			res = append(res, d)
		}
	}
	return res
}

func invoke(t *testing.T, r *http.Request) *http.Response {
	t.Helper()
	resp, err := cfg.httpClient.Do(r)
	require.Nil(t, err, "not nil error = %v", err)
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func checkCode(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		b, _ := io.ReadAll(resp.Body)
		require.Equal(t, expected, resp.StatusCode, string(b))
	}
}

func decode(t *testing.T, resp *http.Response, to interface{}) {
	t.Helper()
	require.Nil(t, json.NewDecoder(resp.Body).Decode(to))
}
