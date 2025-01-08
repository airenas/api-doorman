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
	"github.com/joho/godotenv"
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
	_ = godotenv.Load("../../../.env")
	cfg.url = os.Getenv("ADMIN_URL")
	if cfg.url == "" {
		log.Fatal("FAIL: no ADMIN_URL set")
	}
	cfg.httpClient = &http.Client{Timeout: time.Second * 60} // use bigger for debug

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

	in := &api.CreateInput{ID: uuid.NewString(), OperationID: uuid.NewString(), Service: "test", Credits: 100}
	key := newKeyInput(t, in)
	assert.NotEmpty(t, key.Key)

	resp := invoke(t, newRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusBadRequest)

	resp = invoke(t, newRequest(t, http.MethodPost, "/key",
		api.CreateInput{ID: uuid.NewString(), OperationID: in.OperationID,
			Service: "test", Credits: 100}))
	checkCode(t, resp, http.StatusBadRequest)
}

func TestCreate_OKSaveRequests(t *testing.T) {
	t.Parallel()

	in := &api.CreateInput{ID: uuid.NewString(), OperationID: uuid.NewString(), Service: "test", Credits: 100, SaveRequests: true}
	key := newKeyInput(t, in)
	assert.NotEmpty(t, key.Key)

	res := getKeyInfo(t, in.ID)
	assert.Equal(t, in.Credits, res.TotalCredits)
	assert.Equal(t, "", res.Key)
	assert.True(t, res.SaveRequests)
}

func TestAddCredits(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	resG := getKeyInfo(t, key.ID)
	assert.Equal(t, 100.0, resG.TotalCredits)

	addCredits(t, key, 1000)
	resG = getKeyInfo(t, key.ID)
	assert.Equal(t, 1100.0, resG.TotalCredits)

	addCredits(t, key, -200)
	resG = getKeyInfo(t, key.ID)
	assert.Equal(t, 900.0, resG.TotalCredits)
}

func TestAddCredits_FailLimit(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	resp := addCreditsResp(t, key, -1000, "")
	checkCode(t, resp, http.StatusBadRequest)
}

func TestAddCredits_OKSameOpID(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	id := uuid.NewString()
	resp := addCreditsResp(t, key, 1000, id)
	checkCode(t, resp, http.StatusOK)

	resp = addCreditsResp(t, key, 1000, id)
	checkCode(t, resp, http.StatusOK)

	resp = addCreditsResp(t, key, 1000, id)
	checkCode(t, resp, http.StatusOK)

	resG := getKeyInfo(t, key.ID)
	assert.Equal(t, 1100.0, resG.TotalCredits)
}

func TestGet(t *testing.T) {
	t.Parallel()
	in := api.CreateInput{ID: uuid.NewString(), OperationID: uuid.NewString(), Service: "test", Credits: 100}
	checkCode(t, invoke(t, newRequest(t, http.MethodPost, "/key", in)), http.StatusCreated)

	res := getKeyInfo(t, in.ID)
	assert.Equal(t, in.Credits, res.TotalCredits)
	assert.Equal(t, "", res.Key)
}

func TestUpdate_Disabled(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	res := getKeyInfo(t, key.ID)
	assert.False(t, res.Disabled)

	key = update(t, key.ID, map[string]interface{}{"disabled": true})
	assert.True(t, key.Disabled)
	res = getKeyInfo(t, key.ID)
	assert.True(t, res.Disabled)

	key = update(t, key.ID, map[string]interface{}{"disabled": false})
	assert.False(t, key.Disabled)
	res = getKeyInfo(t, key.ID)
	assert.False(t, res.Disabled)
}

func TestUpdate_FailDisabled(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	resp := updateResp(t, key.ID, map[string]interface{}{"disabled": "olia"})
	checkCode(t, resp, http.StatusBadRequest)
}

func TestUpdate_ValidTo(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	now := time.Now()

	res := getKeyInfo(t, key.ID)
	assert.Greater(t, *res.ValidTo, now)

	key = update(t, key.ID, map[string]interface{}{"validTo": now.Add(time.Hour)})
	assert.Equal(t, now.Add(time.Hour).Unix(), key.ValidTo.Unix())
	res = getKeyInfo(t, key.ID)
	assert.Equal(t, now.Add(time.Hour).Unix(), res.ValidTo.Unix())
}

func TestUpdate_All(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	now := time.Now()

	key = update(t, key.ID, map[string]interface{}{"validTo": now.Add(time.Hour), "disabled": true, "IPWhiteList": "192.123.123.1/32", "description": "olia"})
	assert.Equal(t, now.Add(time.Hour).Unix(), key.ValidTo.Unix())
	res := getKeyInfo(t, key.ID)
	assert.Equal(t, now.Add(time.Hour).Unix(), res.ValidTo.Unix())
	assert.True(t, res.Disabled)
	assert.Equal(t, "192.123.123.1/32", res.IPWhiteList)
	assert.Equal(t, "olia", res.Description)
}

func TestUpdate_FailValidTo(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	resp := updateResp(t, key.ID, map[string]interface{}{"validTo": "olia"})
	checkCode(t, resp, http.StatusBadRequest)
}

func TestUpdate_FailIPWhiteList(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	resp := updateResp(t, key.ID, map[string]interface{}{"IPWhiteList": "olia"})
	checkCode(t, resp, http.StatusBadRequest)
}

func TestUpdate_Fail(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	resp := updateResp(t, key.ID, map[string]interface{}{})
	checkCode(t, resp, http.StatusBadRequest)

	resp = updateResp(t, key.ID+"1", map[string]interface{}{"disabled": true, "validTo": time.Now().Add(time.Hour)})
	checkCode(t, resp, http.StatusBadRequest)
}

func TestGet_ReturnsKey(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	res := getKeyInfo(t, key.ID)
	assert.Empty(t, res.Key)
}

func TestGet_ReturnsKeyID(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	resp := invoke(t, newRequest(t, http.MethodPost, "/keyID", api.Key{Key: key.Key}))
	checkCode(t, resp, http.StatusOK)
	resKey := api.KeyID{}
	decode(t, resp, &resKey)

	assert.Equal(t, key.ID, resKey.ID)
	assert.Equal(t, key.Service, resKey.Service)
}

func TestChangeKey_OK(t *testing.T) {
	t.Parallel()

	key := newKey(t)
	old := key.Key
	
	resp := invoke(t, newRequest(t, http.MethodPost, fmt.Sprintf("/key/%s/change", key.ID), nil))
	checkCode(t, resp, http.StatusOK)
	nKey := api.Key{}
	decode(t, resp, &nKey)
	assert.NotEmpty(t, nKey.Key)
	assert.NotEqual(t, old, nKey.Key)

	resp = invoke(t, newRequest(t, http.MethodPost, "/keyID", api.Key{Key: old}))
	checkCode(t, resp, http.StatusBadRequest)

	resp = invoke(t, newRequest(t, http.MethodPost, "/keyID", api.Key{Key: nKey.Key}))
	checkCode(t, resp, http.StatusOK)
	resKey := api.KeyID{}
	decode(t, resp, &resKey)

	assert.Equal(t, key.ID, resKey.ID)
	assert.Equal(t, key.Service, resKey.Service)
}

func TestChangeKey_Fail(t *testing.T) {
	t.Parallel()

	resp := invoke(t, newRequest(t, http.MethodPost, fmt.Sprintf("/key/%s/change", "olia"), nil))
	checkCode(t, resp, http.StatusBadRequest)
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

func getKeyInfo(t *testing.T, s string) *api.Key {
	t.Helper()

	resp := invoke(t, newRequest(t, http.MethodGet, "/key/"+s, nil))
	checkCode(t, resp, http.StatusOK)
	res := api.Key{}
	decode(t, resp, &res)
	return &res
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

func addCredits(t *testing.T, key *api.Key, quota float64) *api.Key {
	t.Helper()

	resp := addCreditsResp(t, key, quota, uuid.NewString())
	checkCode(t, resp, http.StatusOK)
	res := api.Key{}
	decode(t, resp, &res)
	return &res
}

func addCreditsResp(t *testing.T, key *api.Key, quota float64, opID string) *http.Response {
	t.Helper()

	if opID == "" {
		opID = uuid.NewString()
	}

	in := api.CreditsInput{OperationID: opID, Credits: quota, Msg: "test"}
	return invoke(t, newRequest(t, http.MethodPatch, fmt.Sprintf("/key/%s/credits", key.ID), in))
}

func update(t *testing.T, id string, in map[string]interface{}) *api.Key {
	t.Helper()

	resp := updateResp(t, id, in)
	checkCode(t, resp, http.StatusOK)
	res := api.Key{}
	decode(t, resp, &res)
	return &res
}

func updateResp(t *testing.T, id string, in map[string]interface{}) *http.Response {
	t.Helper()

	return invoke(t, newRequest(t, http.MethodPatch, fmt.Sprintf("/key/%s", id), in))
}

func newKey(t *testing.T) *api.Key {
	t.Helper()

	return newKeyInput(t, &api.CreateInput{ID: uuid.NewString(), OperationID: uuid.NewString(), Service: "test", Credits: 100})
}

func newKeyInput(t *testing.T, in *api.CreateInput) *api.Key {
	t.Helper()

	resp := invoke(t, newRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusCreated)
	res := api.Key{}
	decode(t, resp, &res)
	assert.NotEmpty(t, res.Key)
	return &res
}
