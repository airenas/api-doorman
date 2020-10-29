package admin

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks"
)

var keyCreatorMock *mocks.MockKeyCreator

var keyRetrieverMock *mocks.MockKeyRetriever

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	keyCreatorMock = mocks.NewMockKeyCreator()
	keyRetrieverMock = mocks.NewMockKeyRetriever()
	//	pegomock.When(recognizerMapMock.Get(pegomock.AnyString())).ThenReturn("recID", nil)
}

func TestWrongPath(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("GET", "/invalid", nil)
	testCode(t, req, 404)
}

func TestKeyList(t *testing.T) {
	initTest(t)
	pegomock.When(keyRetrieverMock.List()).ThenReturn([]*adminapi.Key{}, nil)
	req := httptest.NewRequest("GET", "/key-list", nil)
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "[]\n", string(bytes))
}

func TestKeyList_Returns(t *testing.T) {
	initTest(t)
	pegomock.When(keyRetrieverMock.List()).ThenReturn([]*adminapi.Key{&adminapi.Key{Key: "olia"}}, nil)
	req := httptest.NewRequest("GET", "/key-list", nil)
	resp := testCode(t, req, 200)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), `"key":"olia"`)
}

func TestKeyList_Fail(t *testing.T) {
	initTest(t)
	pegomock.When(keyRetrieverMock.List()).ThenReturn(nil, errors.New("olia"))
	req := httptest.NewRequest("GET", "/key-list", nil)
	testCode(t, req, 500)
}

func newTestRouter() *mux.Router {
	return NewRouter(newTestData())
}

func newTestData() *Data {
	res := &Data{KeySaver: keyCreatorMock,
		KeyGetter: keyRetrieverMock}
	return res
}

func testCode(t *testing.T, req *http.Request, code int) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	newTestRouter().ServeHTTP(resp, req)
	assert.Equal(t, code, resp.Code)
	return resp
}
