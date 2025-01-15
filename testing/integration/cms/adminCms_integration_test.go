package cms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/model/permission"
	"github.com/airenas/api-doorman/internal/pkg/test"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
	"github.com/airenas/api-doorman/testing/integration"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type config struct {
	url        string
	doormanURL string
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
	cfg.doormanURL = os.Getenv("DOORMAN_URL")
	if cfg.doormanURL == "" {
		log.Fatal("FAIL: no DOORMAN_URL set")
	}
	cfg.httpClient = &http.Client{Timeout: time.Second * 60} // use bigger for debug

	tCtx, cf := context.WithTimeout(context.Background(), time.Second*20)
	defer cf()
	test.WaitForOpenOrFail(tCtx, cfg.url)
	db, err := integration.NewDB()
	if err != nil {
		log.Fatal("FAIL: no DB")
	}
	defer db.Close()
	cfg.db = db

	os.Exit(m.Run())
}

func TestLive(t *testing.T) {
	t.Parallel()
	checkCode(t, invoke(t, newRequest(t, http.MethodGet, "/live", nil)), http.StatusOK)
}

func TestCreate(t *testing.T) {
	t.Parallel()

	in := &api.CreateInput{ID: ulid.Make().String(), OperationID: ulid.Make().String(), Service: "test", Credits: 100}
	key := newKeyInput(t, in)
	assert.NotEmpty(t, key.Key)

	resp := invoke(t, newRequest(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusConflict)

	resp = invoke(t, newRequest(t, http.MethodPost, "/key",
		api.CreateInput{ID: ulid.Make().String(), OperationID: in.OperationID,
			Service: "test", Credits: 200}))
	checkCode(t, resp, http.StatusBadRequest)
}

func TestCreate_OKSaveRequests(t *testing.T) {
	t.Parallel()

	in := &api.CreateInput{ID: ulid.Make().String(), OperationID: ulid.Make().String(), Service: "test", Credits: 100, SaveRequests: true}
	key := newKeyInput(t, in)
	assert.NotEmpty(t, key.Key)

	res := getKeyInfo(t, in.ID)
	assert.Equal(t, in.Credits, res.TotalCredits)
	assert.Equal(t, "", res.Key)
	assert.True(t, res.SaveRequests)
}

func TestCreate_OKAllfields(t *testing.T) {
	t.Parallel()

	to := time.Now().Add(time.Hour)
	in := &api.CreateInput{ID: ulid.Make().String(), OperationID: ulid.Make().String(),
		Service: "test", Credits: 100, Description: "olia", SaveRequests: true,
		IPWhiteList: "1.1.1.1/32", Disabled: true, ValidTo: &to}
	key := newKeyInput(t, in)
	assert.NotEmpty(t, key.Key)
	assert.Equal(t, to.Unix(), key.ValidTo.Unix())
	assert.Equal(t, "olia", key.Description)
	assert.Equal(t, "1.1.1.1/32", key.IPWhiteList)
	assert.True(t, key.Disabled)
	assert.True(t, key.SaveRequests)
}

func TestCreate_FailNoAuth(t *testing.T) {
	t.Parallel()

	in := &api.CreateInput{ID: ulid.Make().String(), OperationID: ulid.Make().String(), Service: "test", Credits: 100, SaveRequests: true}
	resp := invoke(t, newRequestNoAuth(t, http.MethodPost, "/key", in))
	checkCode(t, resp, http.StatusUnauthorized)
}

func TestCreate_FailNoProject(t *testing.T) {
	t.Parallel()

	key := newAdminKey(t, &integration.InsertAdminParams{
		Projects:    []string{"tts"},
		Permissions: []string{},
		MaxLimit:    1000,
		MaxValidTo:  time.Now().AddDate(1, 0, 0),
	})
	in := &api.CreateInput{ID: ulid.Make().String(), OperationID: ulid.Make().String(), Service: "test", Credits: 100, SaveRequests: true}
	resp := invoke(t, newRequestWithAuth(t, http.MethodPost, "/key", in, key))
	checkCode(t, resp, http.StatusForbidden)
}

func newAdminKey(t *testing.T, in *integration.InsertAdminParams) string {
	key := ulid.Make().String()
	inH := adminapi.KeyIn{Key: key}
	resp := invoke(t, newRequest(t, http.MethodPost, "/hash", inH))
	checkCode(t, resp, http.StatusOK)
	b, _ := io.ReadAll(resp.Body)
	in.KeyHash = string(b)

	integration.InsertAdmin(t, cfg.db, in)
	return key
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

	id := ulid.Make().String()
	resp := addCreditsResp(t, key, 1000, id)
	checkCode(t, resp, http.StatusOK)

	resp = addCreditsResp(t, key, 1000, id)
	checkCode(t, resp, http.StatusOK)

	resp = addCreditsResp(t, key, 2000, id)
	checkCode(t, resp, http.StatusBadRequest)

	resG := getKeyInfo(t, key.ID)
	assert.Equal(t, 1100.0, resG.TotalCredits)
}

func TestAddCredits_FailMaxLimit(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	resp := addCreditsResp(t, key, 100000000000, "")
	checkCode(t, resp, http.StatusBadRequest)
}

func TestAddCredits_FailNoAuth(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	in := api.CreditsInput{OperationID: ulid.Make().String(), Credits: 10, Msg: "test"}
	resp := invoke(t, newRequestNoAuth(t, http.MethodPatch, fmt.Sprintf("/key/%s/credits", key.ID), in))
	checkCode(t, resp, http.StatusUnauthorized)
}

func TestAddCredits_FailOther(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	lKey := newAdminKey(t, &integration.InsertAdminParams{
		Projects:    []string{"test"},
		Permissions: []string{},
		MaxLimit:    1000,
		MaxValidTo:  time.Now().AddDate(1, 0, 0),
	})

	in := api.CreditsInput{OperationID: ulid.Make().String(), Credits: 10, Msg: "test"}
	resp := invoke(t, newRequestWithAuth(t, http.MethodPatch, fmt.Sprintf("/key/%s/credits", key.ID), in, lKey))
	checkCode(t, resp, http.StatusForbidden)
}

func TestGet(t *testing.T) {
	t.Parallel()
	key := newKey(t)

	res := getKeyInfo(t, key.ID)
	assert.Equal(t, "", res.Key)
}

func TestGet_FailNoAuth(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	resp := invoke(t, newRequestNoAuth(t, http.MethodGet, "/key/"+key.ID, nil))
	checkCode(t, resp, http.StatusUnauthorized)
}

func TestGet_FailOther(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	lKey := newAdminKey(t, &integration.InsertAdminParams{
		Projects:    []string{"test"},
		Permissions: []string{},
		MaxLimit:    1000,
		MaxValidTo:  time.Now().AddDate(1, 0, 0),
	})

	resp := invoke(t, newRequestWithAuth(t, http.MethodGet, "/key/"+key.ID, nil, lKey))
	checkCode(t, resp, http.StatusForbidden)
}

func TestGet_OKSuperUser(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	lKey := newAdminKey(t, &integration.InsertAdminParams{
		Projects:    []string{"test"},
		Permissions: []string{permission.Everything.String()},
		MaxLimit:    1000,
		MaxValidTo:  time.Now().AddDate(1, 0, 0),
	})

	resp := invoke(t, newRequestWithAuth(t, http.MethodGet, "/key/"+key.ID, nil, lKey))
	checkCode(t, resp, http.StatusOK)
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

func TestUpdate_FailNoAuth(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	in := map[string]interface{}{"disabled": true}
	resp := invoke(t, newRequestNoAuth(t, http.MethodPatch, fmt.Sprintf("/key/%s", key.ID), in))
	checkCode(t, resp, http.StatusUnauthorized)
}

func TestUpdate_FailOtherUser(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	lKey := newAdminKey(t, &integration.InsertAdminParams{
		Projects:    []string{"test"},
		Permissions: []string{},
		MaxLimit:    1000,
		MaxValidTo:  time.Now().AddDate(1, 0, 0),
	})

	in := map[string]interface{}{"disabled": true}
	resp := invoke(t, newRequestWithAuth(t, http.MethodPatch, fmt.Sprintf("/key/%s", key.ID), in, lKey))
	checkCode(t, resp, http.StatusForbidden)
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

func TestGetID_NoAuth(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	resp := invoke(t, newRequestNoAuth(t, http.MethodPost, "/keyID", api.Key{Key: key.Key}))
	checkCode(t, resp, http.StatusUnauthorized)
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

func TestChangeKey_FailNonManual(t *testing.T) {
	t.Parallel()

	id := integration.InsertIPKey(t, cfg.db, "test")

	resp := invoke(t, newRequest(t, http.MethodPost, fmt.Sprintf("/key/%s/change", id), nil))
	checkCode(t, resp, http.StatusBadRequest)
}

func TestChangeKey_FailNoAuth(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	resp := invoke(t, newRequestNoAuth(t, http.MethodPost, fmt.Sprintf("/key/%s/change", key.ID), nil))
	checkCode(t, resp, http.StatusUnauthorized)
}

func TestChangeKey_Fail(t *testing.T) {
	t.Parallel()

	resp := invoke(t, newRequest(t, http.MethodPost, fmt.Sprintf("/key/%s/change", "olia"), nil))
	checkCode(t, resp, http.StatusBadRequest)
}

func TestKey_NotFound(t *testing.T) {
	t.Parallel()
	checkCode(t, invoke(t, newRequest(t, http.MethodGet,
		fmt.Sprintf("/key/%s", ulid.Make().String()), nil)), http.StatusBadRequest)
}

func TestUsage_Empty(t *testing.T) {
	t.Parallel()

	key := newKey(t)
	now := time.Now()

	resp := invoke(t, newRequest(t, http.MethodGet, fmt.Sprintf("/key/%s/usage?from=%s&to=%s&full=1", key.ID,
		test.TimeToQueryStr(now.Add(-time.Hour)), test.TimeToQueryStr(now.Add(time.Second))), nil))
	checkCode(t, resp, http.StatusOK)
	res := api.Usage{}
	decode(t, resp, &res)
	assert.Equal(t, 0, res.RequestCount)
	assert.Equal(t, 0.0, res.FailedCredits)
	assert.Equal(t, 0.0, res.UsedCredits)
	assert.Len(t, res.Logs, 0)
}

func TestUsage_OK(t *testing.T) {
	t.Parallel()

	key := newKey(t)
	addCredits(t, key, 1000)
	now := time.Now()

	for i := 0; i < 10; i++ {
		newCallService(t, key.Key, 50, http.StatusOK)
	}

	resp := invoke(t, newRequest(t, http.MethodGet, fmt.Sprintf("/key/%s/usage?from=%s&to=%s&full=1", key.ID,
		test.TimeToQueryStr(now.Add(-time.Hour)), test.TimeToQueryStr(now.Add(time.Second))), nil))
	checkCode(t, resp, http.StatusOK)
	res := api.Usage{}
	decode(t, resp, &res)
	assert.Equal(t, 10, res.RequestCount)
	assert.Equal(t, 0.0, res.FailedCredits)
	assert.Equal(t, 500.0, res.UsedCredits)
	assert.Len(t, res.Logs, 10)
}

func TestUsage_OKWithFailures(t *testing.T) {
	t.Parallel()

	key := newKey(t)
	addCredits(t, key, 400)
	now := time.Now()

	for i := 0; i < 10; i++ {
		newCallService(t, key.Key, 50, http.StatusOK)
	}

	for i := 0; i < 20; i++ {
		newCallService(t, key.Key, 10, http.StatusForbidden)
	}

	resp := invoke(t, newRequest(t, http.MethodGet, fmt.Sprintf("/key/%s/usage?from=%s&to=%s&full=1", key.ID,
		test.TimeToQueryStr(now.Add(-time.Hour)), test.TimeToQueryStr(now.Add(time.Second))), nil))
	checkCode(t, resp, http.StatusOK)
	res := api.Usage{}
	decode(t, resp, &res)
	assert.Equal(t, 30, res.RequestCount)
	assert.Equal(t, 200.0, res.FailedCredits)
	assert.Equal(t, 500.0, res.UsedCredits)
	assert.Len(t, res.Logs, 30)
}

func TestUsage_OKNoLog(t *testing.T) {
	t.Parallel()

	key := newKey(t)
	now := time.Now()

	for i := 0; i < 5; i++ {
		newCallService(t, key.Key, 10, http.StatusOK)
	}

	resp := invoke(t, newRequest(t, http.MethodGet, fmt.Sprintf("/key/%s/usage?from=%s&to=%s&full=0", key.ID,
		test.TimeToQueryStr(now.Add(-time.Hour)), test.TimeToQueryStr(now.Add(time.Second))), nil))
	checkCode(t, resp, http.StatusOK)
	res := api.Usage{}
	decode(t, resp, &res)
	assert.Equal(t, 5, res.RequestCount)
	assert.Equal(t, 0.0, res.FailedCredits)
	assert.Equal(t, 50.0, res.UsedCredits)
	assert.Len(t, res.Logs, 0)
}

func TestUsage_FailNoAuth(t *testing.T) {
	t.Parallel()

	key := newKey(t)
	now := time.Now()

	resp := invoke(t, newRequestNoAuth(t, http.MethodGet, fmt.Sprintf("/key/%s/usage?from=%s&to=%s&full=1", key.ID,
		test.TimeToQueryStr(now.Add(-time.Hour)), test.TimeToQueryStr(now.Add(time.Second))), nil))
	checkCode(t, resp, http.StatusUnauthorized)
}

func TestStats_OK(t *testing.T) {
	t.Parallel()

	key := newKey(t)
	addCredits(t, key, 1000)

	for i := 0; i < 10; i++ {
		newCallService(t, key.Key, 50, http.StatusOK)
	}

	// integration.RefreshView(t, cfg.db, "daily_logs")
	// integration.RefreshView(t, cfg.db, "monthly_logs")

	resp := invoke(t, newRequest(t, http.MethodGet, fmt.Sprintf("/key/%s/stats?type=daily", key.ID), nil))
	checkCode(t, resp, http.StatusOK)
	res := []*api.Bucket{}
	decode(t, resp, &res)
	require.Len(t, res, 1)
	assert.Equal(t, 10, res[0].RequestCount)
	assert.Equal(t, 500.0, res[0].UsedQuota)
	assert.Equal(t, 0.0, res[0].FailedQuota)
	assert.Equal(t, 0, res[0].FailedRequests)

	resp = invoke(t, newRequest(t, http.MethodGet, fmt.Sprintf("/key/%s/stats?type=monthly", key.ID), nil))
	checkCode(t, resp, http.StatusOK)
	res = []*api.Bucket{}
	decode(t, resp, &res)
	require.Len(t, res, 1)
	assert.Equal(t, 10, res[0].RequestCount)
	assert.Equal(t, 500.0, res[0].UsedQuota)
	assert.Equal(t, 0.0, res[0].FailedQuota)
	assert.Equal(t, 0, res[0].FailedRequests)
}

func TestStats_OKWithFailures(t *testing.T) {
	t.Parallel()

	key := newKey(t)
	addCredits(t, key, 400)

	for i := 0; i < 10; i++ {
		newCallService(t, key.Key, 50, http.StatusOK)
	}

	for i := 0; i < 20; i++ {
		newCallService(t, key.Key, 10, http.StatusForbidden)
	}

	// integration.RefreshView(t, cfg.db, "daily_logs")
	// integration.RefreshView(t, cfg.db, "monthly_logs")

	resp := invoke(t, newRequest(t, http.MethodGet, fmt.Sprintf("/key/%s/stats?type=daily", key.ID), nil))
	checkCode(t, resp, http.StatusOK)
	res := []*api.Bucket{}
	decode(t, resp, &res)
	require.Len(t, res, 1)
	assert.Equal(t, 30, res[0].RequestCount)
	assert.Equal(t, 500.0, res[0].UsedQuota)
	assert.Equal(t, 200.0, res[0].FailedQuota)
	assert.Equal(t, 20, res[0].FailedRequests)

	resp = invoke(t, newRequest(t, http.MethodGet, fmt.Sprintf("/key/%s/stats?type=monthly", key.ID), nil))
	checkCode(t, resp, http.StatusOK)
	res = []*api.Bucket{}
	decode(t, resp, &res)
	require.Len(t, res, 1)
	assert.Equal(t, 30, res[0].RequestCount)
	assert.Equal(t, 500.0, res[0].UsedQuota)
	assert.Equal(t, 200.0, res[0].FailedQuota)
	assert.Equal(t, 20, res[0].FailedRequests)
}

func TestStats_OKDate(t *testing.T) {
	t.Parallel()

	key := newKey(t)
	now := time.Now()

	addCredits(t, key, 1000)

	for i := 0; i < 5; i++ {
		newCallService(t, key.Key, 10, http.StatusOK)
	}

	// integration.RefreshView(t, cfg.db, "daily_logs")

	resp := invoke(t, newRequest(t, http.MethodGet, fmt.Sprintf("/key/%s/stats?type=daily&from=%s&to=%s", key.ID,
		test.TimeToQueryStr(now.AddDate(0, 0, -2)), test.TimeToQueryStr(now.AddDate(0, 0, 1))), nil))
	checkCode(t, resp, http.StatusOK)
	res := []*api.Bucket{}
	decode(t, resp, &res)
	require.Len(t, res, 1)
	assert.Equal(t, 5, res[0].RequestCount)
	assert.Equal(t, 50.0, res[0].UsedQuota)
	assert.Equal(t, 0.0, res[0].FailedQuota)
	assert.Equal(t, 0, res[0].FailedRequests)

	resp = invoke(t, newRequest(t, http.MethodGet, fmt.Sprintf("/key/%s/stats?type=monthly&from=%s&to=%s", key.ID,
		test.TimeToQueryStr(now.AddDate(0, 0, -2)), test.TimeToQueryStr(now.AddDate(0, 0, 1))), nil))
	checkCode(t, resp, http.StatusOK)
}

func TestStats_FailNoAuth(t *testing.T) {
	t.Parallel()

	key := newKey(t)

	resp := invoke(t, newRequestNoAuth(t, http.MethodGet, fmt.Sprintf("/key/%s/stats?type=monthly", key.ID), nil))
	checkCode(t, resp, http.StatusUnauthorized)
}

func TestStats_FailOtherUser(t *testing.T) {
	t.Parallel()

	key := newKey(t)
	lKey := newAdminKey(t, &integration.InsertAdminParams{
		Projects:    []string{"test"},
		Permissions: []string{},
		MaxLimit:    1000,
		MaxValidTo:  time.Now().AddDate(1, 0, 0),
	})

	resp := invoke(t, newRequestWithAuth(t, http.MethodGet, fmt.Sprintf("/key/%s/stats?type=monthly", key.ID), nil, lKey))
	checkCode(t, resp, http.StatusForbidden)
}

func TestStats_OKNonSuperUser(t *testing.T) {
	t.Parallel()

	lKey := newAdminKey(t, &integration.InsertAdminParams{
		Projects:    []string{"test"},
		Permissions: []string{},
		MaxLimit:    1000,
		MaxValidTo:  time.Now().AddDate(1, 0, 0),
	})

	key := newKeyWithAuth(t, lKey)

	resp := invoke(t, newRequestWithAuth(t, http.MethodGet, fmt.Sprintf("/key/%s/stats?type=monthly", key.ID), nil, lKey))
	checkCode(t, resp, http.StatusOK)
}

type testReq struct {
	Text string `json:"text"`
}

func newCallService(t *testing.T, key string, size int, code int) {
	t.Helper()

	inTest := testReq{Text: strings.Repeat("a", size)}
	resp := invoke(t, addAuth(newDRequest(t, http.MethodPost, "/private", inTest), key))
	checkCode(t, resp, code)
}

func getKeyInfo(t *testing.T, s string) *api.Key {
	t.Helper()

	resp := invoke(t, newRequest(t, http.MethodGet, "/key/"+s, nil))
	checkCode(t, resp, http.StatusOK)
	res := api.Key{}
	decode(t, resp, &res)
	return &res
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

	resp := addCreditsResp(t, key, quota, ulid.Make().String())
	checkCode(t, resp, http.StatusOK)
	res := api.Key{}
	decode(t, resp, &res)
	return &res
}

func addCreditsResp(t *testing.T, key *api.Key, quota float64, opID string) *http.Response {
	t.Helper()

	if opID == "" {
		opID = ulid.Make().String()
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

	return newKeyInput(t, &api.CreateInput{ID: ulid.Make().String(), OperationID: ulid.Make().String(), Service: "test", Credits: 100})
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

func newKeyWithAuth(t *testing.T, key string) *api.Key {
	t.Helper()

	return newKeyInputWithAuth(t, &api.CreateInput{ID: ulid.Make().String(), OperationID: ulid.Make().String(), Service: "test", Credits: 100}, key)
}

func newKeyInputWithAuth(t *testing.T, in *api.CreateInput, key string) *api.Key {
	t.Helper()

	resp := invoke(t, newRequestWithAuth(t, http.MethodPost, "/key", in, key))
	checkCode(t, resp, http.StatusCreated)
	res := api.Key{}
	decode(t, resp, &res)
	assert.NotEmpty(t, res.Key)
	return &res
}

func newRequest(t *testing.T, method string, urlSuffix string, body interface{}) *http.Request {
	t.Helper()

	return integration.AddAdmAuth(newRequestFull(t, method, cfg.url+urlSuffix, body))
}

func newRequestWithAuth(t *testing.T, method string, urlSuffix string, body interface{}, key string) *http.Request {
	t.Helper()

	return integration.AddAuth(newRequestFull(t, method, cfg.url+urlSuffix, body), key)
}

func newRequestNoAuth(t *testing.T, method string, urlSuffix string, body interface{}) *http.Request {
	t.Helper()

	return newRequestFull(t, method, cfg.url+urlSuffix, body)
}

func newRequestFull(t *testing.T, method string, url string, body interface{}) *http.Request {
	t.Helper()

	ctx, cf := context.WithTimeout(context.Background(), time.Second*60)
	t.Cleanup(cf)

	req, err := http.NewRequestWithContext(ctx, method, url, mocks.ToReader(body))
	require.Nil(t, err, "not nil error = %v", err)
	if body != nil {
		req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	return req
}

func newDRequest(t *testing.T, method string, urlSuffix string, body interface{}) *http.Request {
	t.Helper()

	return newRequestFull(t, method, cfg.doormanURL+urlSuffix, body)
}

func addAuth(req *http.Request, s string) *http.Request {
	req.Header.Add(echo.HeaderAuthorization, "Key "+s)
	return req
}
