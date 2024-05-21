package service

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/airenas/go-app/pkg/goapp"

	"github.com/airenas/api-doorman/internal/pkg/mongodb"
	"github.com/airenas/api-doorman/internal/pkg/test/mocks"

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
    skipFirstURL: http://tts:8000/{{rID}}
  rateLimit:
    window: 3m
    default: 1002
    url: redis:6379
  prefixURL: /start
  method: POST
  cleanHeaders: tts-one,tts-two
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
	assert.Contains(t, h.Info(), "FillOutHeader")
	assert.Contains(t, h.Info(), "FillKeyHeader")
	assert.Contains(t, h.Info(), "FillRequestIDHeader(db:test)")
	assert.Contains(t, h.Info(), "RateLimitValidate(1002, RedisRateLimiter(redis:6379, 180))")
	assert.Contains(t, h.Info(), "CleanHeader ([TTS-ONE TTS-TWO])")
	assert.Contains(t, h.Info(), "SkipFirstQuota(rID)")
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

func TestQuotaHandleToTxtFile(t *testing.T) {
	h, err := NewHandler("tts", newTestC(t, `
tts:
  backend: http://olia.lt
  type: quota
  db: test
  quota:
    type: toTxtFile
    service: http://olia/ser
    field: fieldText
    default: 100
  prefixURL: /start
  method: POST
`), newTestProvider(t))
	assert.NotNil(t, h)
	assert.Nil(t, err)
	assert.Contains(t, h.Info(), "ToTextAndQuota(fieldText)")
}

func TestQuotaHandleToTxtFile_Fail(t *testing.T) {
	_, err := NewHandler("tts", newTestC(t, `
tts:
  backend: http://olia.lt
  type: quota
  db: test
  quota:
    type: toTxtFile
    service: http://olia/ser
    field: 
    default: 100
  prefixURL: /start
  method: POST
`), newTestProvider(t))
	assert.NotNil(t, err)
	
	_, err = NewHandler("tts", newTestC(t, `
tts:
  backend: http://olia.lt
  type: quota
  db: test
  quota:
    type: toTxtFile
    service: 
    field: olia
    default: 100
  prefixURL: /start
  method: POST
`), newTestProvider(t))
	assert.NotNil(t, err)
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

func TestQuotaHandler_JSONTTS(t *testing.T) {
	h, err := NewHandler("tts", newTestC(t, `
tts:
  backend: http://olia.lt
  type: quota
  db: test
  quota:
    type: jsonTTS
    discount: 0.85
    default: 100
  prefixURL: /start
  stripPrefix: /start
  method: POST
`), newTestProvider(t))
	assert.NotNil(t, h)
	assert.Nil(t, err)
	assert.Contains(t, h.Info(), "JSONTTSField(text)")
	assert.Contains(t, h.Info(), "JSONTTSAsQuota(discount: 0.8500)")
}

func TestQuotaHandler_JSONTTS_Fail(t *testing.T) {
	_, err := NewHandler("tts", newTestC(t, `
tts:
  backend: http://olia.lt
  type: quota
  db: test
  quota:
    type: jsonTTS
    discount: 1.85
    default: 100
  prefixURL: /start
  stripPrefix: /start
  method: POST
`), newTestProvider(t))
	assert.NotNil(t, err)
}

func TestQuotaHandler_SkipFirstQuota_Fail(t *testing.T) {
	_, err := NewHandler("tts", newTestC(t, `
tts:
  backend: http://olia.lt
  type: quota
  db: test
  quota:
    type: jsonTTS
    default: 100
    skipFirstURL: http://tts;8000/{{}} # expexted {{keyID}}
  prefixURL: /start
  stripPrefix: /start
  method: POST
`), newTestProvider(t))
	assert.NotNil(t, err)
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
  cleanHeaders: tts-one,
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
	assert.Contains(t, h.Info(), "FillOutHeader")
	assert.Contains(t, h.Info(), "FillKeyHeader")
	assert.Contains(t, h.Info(), "FillRequestIDHeader(db:test)")
	assert.Contains(t, h.Info(), "CleanHeader ([TTS-ONE])")
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

func TestSimpleHandler(t *testing.T) {
	h, err := NewHandler("tts", newTestC(t, `
tts:
  backend: http://olia.lt
  type: simple
  db: test
  prefixURL: /start
  stripPrefix: /start
  method: POST
`), newTestProvider(t))
	assert.NotNil(t, h)
	assert.Nil(t, err)
	assert.Contains(t, h.Info(), "StripPrefix(/start)")
}

func TestSimpleHandler_FailQuota(t *testing.T) {
	_, err := NewHandler("tts", newTestC(t, `
tts:
  backend: http://olia.lt
  type: simple
  db: test
  prefixURL: /start
  stripPrefix: /start
  method: POST
  quota:
    type: json
    field: field
    default: 100
`), newTestProvider(t))
	assert.NotNil(t, err)
}

func TestPriority(t *testing.T) {
	assert.Equal(t, -1, (&defaultHandler{}).Priority())
	assert.Equal(t, 5, (&prefixHandler{prefix: "/olia"}).Priority())
	assert.Equal(t, 7, (&prefixHandler{prefix: "/olia/3"}).Priority())
}

func newTestProvider(t *testing.T) mongodb.SProvider {
	return mocks.NewMockSProvider()
}

func newTestC(t *testing.T, configStr string) *viper.Viper {
	v := viper.New()
	v.SetConfigType("yaml")
	goapp.InitEnv(v)
	err := v.ReadConfig(strings.NewReader(configStr))
	assert.Nil(t, err, err)
	return v
}
