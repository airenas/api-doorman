package service

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/airenas/go-app/pkg/goapp"

	"github.com/airenas/api-doorman/internal/pkg/mongodb"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func newTestDefH(h http.Handler) *defaultHandler {
	res := defaultHandler{h: h}
	return &res
}

func newTestQuotaH(h http.Handler, prefix, method string) *prefixHandler {
	res := prefixHandler{h: h, methods: initMethods(method), prefix: prefix}
	return &res
}

func TestDefaultProvider(t *testing.T) {
	h, err := NewHandler("default", newTestC(t, "default:\n  backend: http://olia.lt"), nil)
	assert.NotNil(t, h)
	assert.Nil(t, err)
	hd := h.(*defaultHandler)
	assert.NotNil(t, hd)
	assert.Equal(t, "default", hd.Name())
	assert.Equal(t, "Default handler to 'http://olia.lt'", hd.Info())
	assert.NotNil(t, hd.Handler())
	assert.Equal(t, "default", hd.Name())
	assert.True(t, hd.Valid(nil))
}

func TestDefaultProvider_Fail(t *testing.T) {
	h, err := NewHandler("default", newTestC(t, "default:\n  backend: http://"), nil)
	assert.Nil(t, h)
	assert.NotNil(t, err)
}

func TestNewHandler_Fail(t *testing.T) {
	h, err := NewHandler("default1", newTestC(t, "default1:\n  type: olia"), nil)
	assert.Nil(t, h)
	assert.NotNil(t, err)
	h, err = NewHandler("", newTestC(t, "default:\n  backend: http://olia.lt"), nil)
	assert.Nil(t, h)
	assert.NotNil(t, err)
}

const quotaYaml = `
tts:
  backend: http://olia.lt
  type: quota
  db: test
  quota:
    type: json
    field: field
    default: 100
  prefixURL: /start
  method: POST
`

func TestQuotaHandle(t *testing.T) {
	h, err := NewHandler("tts", newTestC(t, quotaYaml), newTestProvider(t))
	assert.NotNil(t, h)
	assert.Nil(t, err)
	hq := h.(*prefixHandler)
	assert.Equal(t, "tts", hq.Name())
	assert.Contains(t, hq.Info(), "tts handler (POST) to 'http://olia.lt', prefix: /start")
	assert.NotNil(t, hq.Handler())
	assert.Equal(t, "tts", hq.Name())
	assert.True(t, hq.Valid(httptest.NewRequest("POST", "/start", nil)))
	assert.Contains(t, h.Info(), "FillHeader")
}

func TestQuotaHandleAudio(t *testing.T) {
	h, err := NewHandler("tts", newTestC(t, `
tts:
  backend: http://olia.lt
  type: quota
  db: test
  quota:
    type: audioDuration
    service: http://olia/ser
    field: field
    default: 100
  prefixURL: /start
  method: POST
`), newTestProvider(t))
	assert.NotNil(t, h)
	assert.Nil(t, err)
	hq := h.(*prefixHandler)
	assert.Equal(t, "tts", hq.Name())
	assert.Contains(t, hq.Info(), "tts handler (POST) to 'http://olia.lt', prefix: /start")
	assert.NotNil(t, hq.Handler())
	assert.Equal(t, "tts", hq.Name())
	assert.True(t, hq.Valid(httptest.NewRequest("POST", "/start", nil)))
}

func TestQuotaHandler_InitStrip(t *testing.T) {
	h, err := NewHandler("tts", newTestC(t, `
tts:
  backend: http://olia.lt
  type: quota
  db: test
  quota:
    type: audioDuration
    service: http://olia/ser
    field: field
    default: 100
  prefixURL: /start
  stripPrefix: /start
  method: POST
`), newTestProvider(t))
	assert.NotNil(t, h)
	assert.Nil(t, err)
	assert.Contains(t, h.Info(), "StripPrefix(/start)")
}

func TestQuotaHandler_NoInitStrip(t *testing.T) {
	h, err := NewHandler("tts", newTestC(t, `
tts:
  backend: http://olia.lt
  type: quota
  db: test
  quota:
    type: audioDuration
    service: http://olia/ser
    field: field
    default: 100
  prefixURL: /start
  method: POST
`), newTestProvider(t))
	assert.NotNil(t, h)
	assert.Nil(t, err)
	assert.NotContains(t, h.Info(), "StripPrefix(/start)")
}

func TestQuotaHandler_Env(t *testing.T) {
	os.Setenv("PROXY_TTS_QUOTA_TYPE", "json")
	os.Setenv("PROXY_TTS_QUOTA_FIELD", "field")
	h, err := NewHandler("tts", goapp.Sub(newTestC(t, `
proxy:
  tts:
    backend: http://olia.lt
    type: quota
    db: test
    prefixURL: /start
    method: POST
`), "proxy"), newTestProvider(t))
	assert.NotNil(t, h)
	assert.Nil(t, err)
}

func TestQuotaHandle_FailType(t *testing.T) {
	os.Setenv("TTS_TYPE", "test")
	defer os.Setenv("TTS_TYPE", "")
	h, err := NewHandler("tts", newTestC(t, quotaYaml), newTestProvider(t))
	assert.Nil(t, h)
	assert.NotNil(t, err)
}

func TestQuotaHandle_FailBacked(t *testing.T) {
	os.Setenv("TTS_BACKEND", " ")
	defer os.Setenv("TTS_BACKEND", "")
	h, err := NewHandler("tts", newTestC(t, quotaYaml), newTestProvider(t))
	assert.Nil(t, h)
	assert.NotNil(t, err)
}

func TestQuotaHandle_FailQuotaType(t *testing.T) {
	os.Setenv("TTS_QUOTA_TYPE", "json1")
	defer os.Setenv("TTS_QUOTA_TYPE", "")
	h, err := NewHandler("tts", newTestC(t, quotaYaml), newTestProvider(t))
	assert.Nil(t, h)
	assert.NotNil(t, err)
}

func TestQuotaHandle_FailQuotaField(t *testing.T) {
	os.Setenv("TTS_QUOTA_FIELD", " ")
	defer os.Setenv("TTS_QUOTA_FIELD", "")
	h, err := NewHandler("tts", newTestC(t, quotaYaml), newTestProvider(t))
	assert.Nil(t, h)
	assert.NotNil(t, err)
}

func TestQuotaHandle_FailDB(t *testing.T) {
	os.Setenv("TTS_DB", " ")
	defer os.Setenv("TTS_DB", "")
	h, err := NewHandler("tts", newTestC(t, quotaYaml), newTestProvider(t))
	assert.Nil(t, h)
	assert.NotNil(t, err)
}

func TestQuotaHandle_FailDurationService(t *testing.T) {
	os.Setenv("TTS_QUOTA_TYPE", "audioDuration")
	defer os.Setenv("TTS_QUOTA_TYPE", "")
	h, err := NewHandler("tts", newTestC(t, quotaYaml), newTestProvider(t))
	assert.Nil(t, h)
	assert.NotNil(t, err)
}

func TestKeyHandler(t *testing.T) {
	h, err := NewHandler("tts", newTestC(t, `
tts:
  backend: http://olia.lt
  type: key
  db: test
  prefixURL: /start
  method: POST
`), newTestProvider(t))
	assert.NotNil(t, h)
	assert.Nil(t, err)
	hq := h.(*prefixHandler)
	assert.Equal(t, "tts", hq.Name())
	assert.Contains(t, hq.Info(), "tts handler (POST) to 'http://olia.lt', prefix: /start")
	assert.NotNil(t, hq.Handler())
	assert.Equal(t, "tts", hq.Name())
	assert.True(t, hq.Valid(httptest.NewRequest("POST", "/start", nil)))
	assert.Contains(t, h.Info(), "FillHeader")
}

func TestKeyHandler_FailNoDB(t *testing.T) {
	h, err := NewHandler("tts", newTestC(t, `
tts:
  backend: http://olia.lt
  type: key
  prefixURL: /start
  method: POST
`), newTestProvider(t))
	assert.Nil(t, h)
	assert.NotNil(t, err)
}

func TestKeyHandler_FailNoBackend(t *testing.T) {
	h, err := NewHandler("tts", newTestC(t, `
tts:
  type: key
  db: db
  prefixURL: /start
  method: POST
`), newTestProvider(t))
	assert.Nil(t, h)
	assert.NotNil(t, err)
}

func TestKeyHandler_FailNoPrefix(t *testing.T) {
	h, err := NewHandler("tts", newTestC(t, `
tts:
  type: key
  backend: http://olia.lt
  db: db
  method: POST
`), newTestProvider(t))
	assert.Nil(t, h)
	assert.NotNil(t, err)
}

func newTestProvider(t *testing.T) *mongodb.SessionProvider {
	res, err := mongodb.NewSessionProvider("mongo://olia")
	assert.Nil(t, err)
	return res
}

func newTestC(t *testing.T, configStr string) *viper.Viper {
	v := viper.New()
	v.SetConfigType("yaml")
	goapp.InitEnv(v)
	err := v.ReadConfig(strings.NewReader(configStr))
	assert.Nil(t, err, err)
	return v
}
