package doorman

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/model/permission"
	"github.com/airenas/api-doorman/internal/pkg/test"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/airenas/api-doorman/testing/integration"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type config struct {
	url        string
	doormanUrl string
	httpClient *http.Client
	db         *sqlx.DB
}

var cfg config

func TestMain(m *testing.M) {
	_ = godotenv.Load("../../../.env")
	cfg.url = os.Getenv("ADMIN_URL")
	if cfg.url == "" {
		log.Fatal("FAIL: no ADMIN_URL set")
	}
	cfg.doormanUrl = os.Getenv("DOORMAN_URL")
	if cfg.doormanUrl == "" {
		log.Fatal("FAIL: no DOORMAN_URL set")
	}
	cfg.httpClient = &http.Client{Timeout: time.Second * 60} // use bigger for debug

	tCtx, cf := context.WithTimeout(context.Background(), time.Second*10)
	defer cf()
	test.WaitForOpenOrFail(tCtx, cfg.url)
	test.WaitForOpenOrFail(tCtx, cfg.doormanUrl)
	db, err := integration.NewDB()
	if err != nil {
		log.Fatal("FAIL: no DB")
	}
	defer db.Close()
	cfg.db = db

	os.Exit(m.Run())
}

func TestLiveAdmin(t *testing.T) {
	t.Parallel()
	checkCode(t, invoke(t, newAdminRequest(t, http.MethodGet, "/live", nil)), http.StatusOK)
}

func TestAccessCreate(t *testing.T) {
	t.Parallel()
	id := ulid.Make().String()
	in := api.CreateInput{ID: id, OperationID: ulid.Make().String(), Service: "test", Credits: 100}
	resp := invoke(t, newAdminRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusCreated)
	res := api.Key{}
	decode(t, resp, &res)
	assert.NotEmpty(t, res.Key)

	resp = invoke(t, newAdminRequest(t, http.MethodGet, fmt.Sprintf("/key/%s", id), in))
	checkCode(t, resp, http.StatusOK)
	res = api.Key{}
	decode(t, resp, &res)
	assert.Equal(t, 100.0, res.TotalCredits)
}

func TestAccessCreate_FailNoAuth(t *testing.T) {
	t.Parallel()

	in := api.CreateInput{ID: ulid.Make().String(), OperationID: ulid.Make().String(), Service: "test", Credits: 100}
	resp := invoke(t, newAdminRequestNoAuth(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusUnauthorized)
}

func TestAccessCreate_FailValidTo(t *testing.T) {
	t.Parallel()

	to := time.Now().AddDate(50, 0, 0)
	in := api.CreateInput{ID: ulid.Make().String(),
		ValidTo: &to,

		OperationID: ulid.Make().String(), Service: "test", Credits: 100}
	resp := invoke(t, newAdminRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusBadRequest)
}

func TestAccessCreate_FailLimits(t *testing.T) {
	t.Parallel()

	in := api.CreateInput{ID: ulid.Make().String(), OperationID: ulid.Make().String(), Service: "test", Credits: -100}
	resp := invoke(t, newAdminRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusBadRequest)

	in = api.CreateInput{ID: ulid.Make().String(), OperationID: ulid.Make().String(), Service: "test", Credits: 10000000000000}
	resp = invoke(t, newAdminRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusBadRequest)
}

func TestAccessCreate_FailWrong(t *testing.T) {
	t.Parallel()

	in := api.CreateInput{ID: ulid.Make().String(), OperationID: ulid.Make().String(), Service: "test", Credits: 100}
	resp := invoke(t, integration.AddAuth(newAdminRequestNoAuth(t, http.MethodPost, "/key", in), "haha"))
	checkCode(t, resp, http.StatusUnauthorized)
}

func TestAccessCreate_FailDuplicate(t *testing.T) {
	t.Parallel()
	id := uuid.NewString()
	in := api.CreateInput{ID: id, OperationID: uuid.NewString(), Service: "test", Credits: 100}
	resp := invoke(t, newAdminRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusCreated)
	res := api.Key{}
	decode(t, resp, &res)
	assert.NotEmpty(t, res.Key)

	resp = invoke(t, newAdminRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusBadRequest)
	in.ID = uuid.NewString()
	resp = invoke(t, newAdminRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusConflict)
}

type testReq struct {
	Text string `json:"text"`
}
type errReq struct {
	Error string `json:"error"`
}

func TestAccessCreate_Used(t *testing.T) {
	t.Parallel()
	id := uuid.NewString()
	in := api.CreateInput{ID: id, OperationID: uuid.NewString(), Service: "test", Credits: 100}
	resp := invoke(t, newAdminRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusCreated)
	res := api.Key{}
	decode(t, resp, &res)
	assert.NotEmpty(t, res.Key)

	inTest := testReq{Text: "olia olia "}
	resp = invoke(t, addAuth(newRequest(t, http.MethodPost, "/private", inTest), "olia"))
	checkCode(t, resp, http.StatusUnauthorized)

	resp = invoke(t, addAuth(newRequest(t, http.MethodPost, "/private", inTest), res.Key))
	checkCode(t, resp, http.StatusOK)

	resp = invoke(t, newAdminRequest(t, http.MethodGet, fmt.Sprintf("/key/%s?returnKey=1", id), nil))
	res = api.Key{}
	decode(t, resp, &res)
	assert.Equal(t, 10.0, res.UsedCredits)
	assert.Equal(t, 0.0, res.FailedCredits)
}

func TestRestore_OK(t *testing.T) {
	t.Parallel()
	id := uuid.NewString()
	in := api.CreateInput{ID: id, OperationID: uuid.NewString(), Service: "test", Credits: 100}
	resp := invoke(t, newAdminRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusCreated)
	res := api.Key{}
	decode(t, resp, &res)
	assert.NotEmpty(t, res.Key)

	inTest := testReq{Text: "olia olia "}
	resp = invoke(t, addAuth(newRequest(t, http.MethodPost, "/private", inTest), res.Key))
	checkCode(t, resp, http.StatusOK)

	resp = invoke(t, newAdminRequest(t, http.MethodGet, fmt.Sprintf("/%s/key/%s?full=1", "test", res.ID), nil))
	checkCode(t, resp, http.StatusOK)
	logs := adminapi.KeyInfoResp{}
	decode(t, resp, &logs)
	require.Equal(t, 1, len(logs.Logs))
	assert.Equal(t, 10.0, logs.Key.QuotaValue)

	resp = invoke(t, newAdminRequest(t, http.MethodPost,
		fmt.Sprintf("/%s/restore/m:%s", "test", logs.Logs[0].RequestID), errReq{Error: "err"}))
	checkCode(t, resp, http.StatusOK)

	resp = invoke(t, newAdminRequest(t, http.MethodPost,
		fmt.Sprintf("/%s/restore/m:%s", "test", logs.Logs[0].RequestID), errReq{Error: "err"}))
	checkCode(t, resp, http.StatusConflict)

	resp = invoke(t, newAdminRequest(t, http.MethodGet, fmt.Sprintf("/key/%s", id), in))
	res = api.Key{}
	decode(t, resp, &res)
	assert.Equal(t, 0.0, res.UsedCredits)
	assert.Equal(t, 10.0, res.FailedCredits)
}

func TestHash_OK(t *testing.T) {
	t.Parallel()

	in := adminapi.KeyIn{Key: "aaa"}
	resp := invoke(t, newAdminRequest(t, http.MethodPost, "/hash", in))
	checkCode(t, resp, http.StatusOK)
}

func TestHash_FailNoAuth(t *testing.T) {
	t.Parallel()

	in := adminapi.KeyIn{Key: "aaa"}
	resp := invoke(t, newAdminRequestNoAuth(t, http.MethodPost, "/hash", in))
	checkCode(t, resp, http.StatusUnauthorized)
}

func TestDefaultUsage(t *testing.T) {
	t.Parallel()

	inTest := testReq{Text: "olia olia "}
	resp := invoke(t, newRequest(t, http.MethodPost, "/private", inTest))
	checkCode(t, resp, http.StatusOK)
}

func TestReset(t *testing.T) {
	t.Parallel()

	integration.ResetSettings(t, cfg.db, "reset-test")
	id1 := integration.InsertIPKey(t, cfg.db, "test")
	id2 := integration.InsertIPKey(t, cfg.db, "test")

	since := time.Now().Add(time.Second)
	resp := invoke(t, newAdminRequest(t, http.MethodPost, fmt.Sprintf("/test/reset?quota=50000&since=%s", test.TimeToQueryStr(since)), nil))
	checkCode(t, resp, http.StatusOK)

	key := getKeyInfo(t, id1)
	assert.Equal(t, 50000.0, key.TotalCredits-key.UsedCredits)

	key = getKeyInfo(t, id2)
	assert.Equal(t, 50000.0, key.TotalCredits-key.UsedCredits)

	// try again
	id3 := integration.InsertIPKey(t, cfg.db, "test")
	integration.ResetSettings(t, cfg.db, "reset-test")
	resp = invoke(t, newAdminRequest(t, http.MethodPost, fmt.Sprintf("/test/reset?quota=40000&since=%s", test.TimeToQueryStr(since)), nil))
	checkCode(t, resp, http.StatusOK)

	// not changed
	key = getKeyInfo(t, id2)
	assert.Equal(t, 50000.0, key.TotalCredits-key.UsedCredits)

	// new
	key = getKeyInfo(t, id3)
	assert.Equal(t, 40000.0, key.TotalCredits-key.UsedCredits)
}

func TestReset_FailNoAuth(t *testing.T) {
	t.Parallel()

	since := time.Now().Add(time.Second)
	resp := invoke(t, newAdminRequestNoAuth(t, http.MethodPost, fmt.Sprintf("/test/reset?quota=50000&since=%s", test.TimeToQueryStr(since)), nil))
	checkCode(t, resp, http.StatusUnauthorized)

	key := ulid.Make().String()
	in := adminapi.KeyIn{Key: key}
	resp = invoke(t, newAdminRequest(t, http.MethodPost, "/hash", in))
	checkCode(t, resp, http.StatusOK)
	b, _ := io.ReadAll(resp.Body)

	integration.InsertAdmin(t, cfg.db, &integration.InsertAdminParams{
		Projects:    []string{"test"},
		KeyHash:     string(b),
		Permissions: []string{},
	})
	resp = invoke(t, newAdminRequestWithAuth(t, http.MethodPost, fmt.Sprintf("/test/reset?quota=50000&since=%s", test.TimeToQueryStr(since)), nil, key))
	checkCode(t, resp, http.StatusForbidden)
}

func TestReset_OKHasPerm(t *testing.T) {
	t.Parallel()

	key := ulid.Make().String()
	in := adminapi.KeyIn{Key: key}
	resp := invoke(t, newAdminRequest(t, http.MethodPost, "/hash", in))
	checkCode(t, resp, http.StatusOK)
	b, _ := io.ReadAll(resp.Body)

	integration.InsertAdmin(t, cfg.db, &integration.InsertAdminParams{
		Projects:    []string{"tts"},
		KeyHash:     string(b),
		Permissions: []string{permission.ResetMonthlyUsage.String()},
	})
	since := time.Now().Add(time.Second)
	resp = invoke(t, newAdminRequestWithAuth(t, http.MethodPost, fmt.Sprintf("/tts/reset?quota=50000&since=%s", test.TimeToQueryStr(since)), nil, key))
	checkCode(t, resp, http.StatusOK)
}

func addAuth(req *http.Request, s string) *http.Request {
	req.Header.Add(echo.HeaderAuthorization, "Key "+s)
	return req
}

func newRequest(t *testing.T, method string, urlSuffix string, body interface{}) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, cfg.doormanUrl+urlSuffix, mocks.ToReader(body))
	require.Nil(t, err, "not nil error = %v", err)
	if body != nil {
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	return req
}

func newAdminRequest(t *testing.T, method string, urlSuffix string, body interface{}) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, cfg.url+urlSuffix, mocks.ToReader(body))
	require.Nil(t, err, "not nil error = %v", err)
	if body != nil {
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	return integration.AddAdmAuth(req)
}

func newAdminRequestWithAuth(t *testing.T, method string, urlSuffix string, body interface{}, key string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, cfg.url+urlSuffix, mocks.ToReader(body))
	require.Nil(t, err, "not nil error = %v", err)
	if body != nil {
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	return integration.AddAuth(req, key)
}

func newAdminRequestNoAuth(t *testing.T, method string, urlSuffix string, body interface{}) *http.Request {
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

func getKeyInfo(t *testing.T, s string) *api.Key {
	t.Helper()

	resp := invoke(t, newAdminRequest(t, http.MethodGet, "/key/"+s, nil))
	checkCode(t, resp, http.StatusOK)
	res := api.Key{}
	decode(t, resp, &res)
	return &res
}
