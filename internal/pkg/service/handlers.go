package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/airenas/api-doorman/internal/pkg/audio"
	"github.com/airenas/api-doorman/internal/pkg/handler"
	"github.com/airenas/api-doorman/internal/pkg/mongodb"
	"github.com/airenas/api-doorman/internal/pkg/text"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

//NewHandler creates handler based on config
func NewHandler(name string, cfg *viper.Viper, ms *mongodb.SessionProvider) (HandlerWrap, error) {
	if name == "default" {
		return newDefaultHandler("default", cfg)
	}
	sType := cfg.GetString(name + ".type")
	if sType == "quota" {
		return newPrQuotaHandler(name, cfg, ms)
	}
	if sType == "simple" {
		return newPrQuotaHandler(name, cfg, ms)
	}
	if sType == "key" {
		return newPrKeyHandler(name, cfg, ms)
	}
	return nil, errors.Errorf("Unknown handler type '%s'", sType)
}

type defaultHandler struct {
	proxyURL string
	name     string
	h        http.Handler
}

func newDefaultHandler(name string, cfg *viper.Viper) (HandlerWrap, error) {
	res := &defaultHandler{}
	res.name = name
	res.proxyURL = cfg.GetString(name + ".backend")
	goapp.Log.Infof("Backend: %s", res.proxyURL)
	url, err := utils.ParseURL(res.proxyURL)
	if err != nil {
		return nil, errors.Wrap(err, "Wrong backendURL")
	}
	res.h = handler.Proxy(url)
	return res, nil
}

func (h *defaultHandler) Handler() http.Handler {
	return h.h
}

func (h *defaultHandler) Info() string {
	return fmt.Sprintf("Default handler to '%s'", h.proxyURL)
}

func (h *defaultHandler) Valid(r *http.Request) bool {
	return true
}

func (h *defaultHandler) Name() string {
	return h.name
}

func (h *defaultHandler) Priority() int {
	return -1
}

type prefixHandler struct {
	prefix   string
	methods  map[string]bool
	proxyURL string
	name     string
	h        http.Handler
}

func newPrQuotaHandler(name string, cfg *viper.Viper, ms *mongodb.SessionProvider) (HandlerWrap, error) {
	res := &prefixHandler{}
	err := initPrefixes(name, cfg, res)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't init prefix for %s", name)
	}
	res.h, err = newQuotaHandler(name, cfg, ms)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init handler")
	}
	return res, nil
}

func initPrefixes(name string, cfg *viper.Viper, res *prefixHandler) error {
	res.name = name
	res.prefix = strings.TrimSpace(strings.ToLower(cfg.GetString(name + ".prefixURL")))
	if res.prefix == "" {
		return errors.New("No prefix")
	}
	res.proxyURL = cfg.GetString(name + ".backend")
	res.methods = initMethods(cfg.GetString(name + ".method"))
	goapp.Log.Infof("PrefixURL: %s", res.prefix)
	return nil
}

func newQuotaHandler(name string, cfg *viper.Viper, ms *mongodb.SessionProvider) (http.Handler, error) {
	if cfg.GetString(name+".backend") == "" {
		return nil, errors.New("No backend")
	}
	goapp.Log.Infof("Backend: %s", cfg.GetString(name+".backend"))
	proxyURL := cfg.GetString(name + ".backend")
	url, err := utils.ParseURL(proxyURL)
	if err != nil {
		return nil, errors.Wrap(err, "Wrong backend")
	}

	dbProvider, err := mongodb.NewDBProvider(ms, cfg.GetString(name+".db"))
	if err != nil {
		return nil, errors.Wrap(err, "No db")
	}
	keysValidator, err := mongodb.NewKeyValidator(dbProvider)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init validator")
	}

	h := handler.FillOutHeader(handler.Proxy(url))
	h = handler.FillHeader(handler.FillKeyHeader(h))
	h, err = addCleanHeader(h, cfg.GetString(name+".cleanHeaderPrefix"))
	if err != nil {
		return nil, errors.Wrap(err, "can't init clean header")
	}
	stripURL := cfg.GetString(name + ".stripPrefix")
	if stripURL != "" {
		h = handler.StripPrefix(h, stripURL)
		goapp.Log.Infof("Strip prefix: %s", stripURL)
	}

	tp := cfg.GetString(name + ".type")
	qt := cfg.GetString(name + ".quota.type")

	if tp == "quota" {
		h = handler.QuotaValidate(h, keysValidator)
		qf := strings.TrimSpace(cfg.GetString(name + ".quota.field"))
		if qt == "json" {
			if qf == "" {
				return nil, errors.New("No field")
			}
			goapp.Log.Infof("Quota extract: %s(%s)", qt, qf)
			h = handler.TakeJSON(handler.JSONAsQuota(h), qf)
		} else if qt == "jsonTTS" {
			goapp.Log.Infof("Quota extract: %s(text)", qt)
			h, err = handler.JSONTTSAsQuota(h, cfg.GetFloat64(name+".quota.discount"))
			if err != nil {
				return nil, errors.Wrap(err, "Can't init jsonQuota handler")
			}
			h = handler.TakeJSONTTS(h)
		} else if qt == "audioDuration" {
			if qf == "" {
				return nil, errors.New("No field")
			}
			dsURL := cfg.GetString(name + ".quota.service")
			ds, err := audio.NewDurationClient(dsURL)
			if err != nil {
				return nil, errors.Wrap(err, "Can't init Duration service")
			}
			goapp.Log.Infof("Duration service: %s", dsURL)
			goapp.Log.Infof("Quota extract: %s(%s) using duration service", qt, qf)
			h = handler.AudioLenQuota(h, qf, ds)
		} else if qt == "toTxtFile" {
			if qf == "" {
				return nil, errors.New("No field")
			}
			dsURL := cfg.GetString(name + ".quota.service")
			ds, err := text.NewExtractor(dsURL)
			if err != nil {
				return nil, errors.Wrap(err, "can't init text extraction service")
			}
			goapp.Log.Infof("Text extraction service: %s", dsURL)
			goapp.Log.Infof("Quota extract: %s(%s) using text extraction service", qt, qf)
			h = handler.ToTextAndQuota(h, qf, ds)
		} else {
			return nil, errors.Errorf("Unknown proxy quota type '%s'", qt)
		}
	} else {
		if qt != "" {
			return nil, errors.Errorf("Quota is not expected for type simple")
		}
		goapp.Log.Infof("No quota validation")
	}

	ls, err := mongodb.NewLogSaver(dbProvider)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init log saver")
	}
	h = handler.LogDB(h, ls)
	hKey := handler.KeyValid(h, keysValidator)
	dl := cfg.GetFloat64(name + ".quota.default")
	if dl > 0 {
		goapp.Log.Infof("Default IP quota: %.f", dl)
		is, err := mongodb.NewIPSaver(dbProvider)
		if err != nil {
			return nil, errors.Wrap(err, "Can't init IP saver")
		}
		hIP := handler.IPAsKey(hKey, newIPSaver(is, dl))
		hKey = handler.KeyValidOrIP(hKey, hIP)
	}
	h = handler.KeyExtract(hKey)

	return h, nil
}

func (h *prefixHandler) Handler() http.Handler {
	return h.h
}

func (h *prefixHandler) Info() string {
	res := fmt.Sprintf("%s handler (%s) to '%s', prefix: %s\n", h.name, keys(h.methods), h.proxyURL, h.prefix)
	return res + handler.GetInfo(handler.LogShitf(""), h.h)
}

func (h *prefixHandler) Priority() int {
	return len(h.prefix)
}

func keys(data map[string]bool) string {
	res := strings.Builder{}
	sep := ""
	for k := range data {
		res.WriteString(sep)
		sep = ", "
		res.WriteString(k)
	}
	return res.String()
}

func (h *prefixHandler) Valid(r *http.Request) bool {
	path := r.URL.Path
	path = strings.ToLower(path)
	return strings.HasPrefix(path, h.prefix) && h.methodOK(r.Method)
}

func (h *prefixHandler) methodOK(m string) bool {
	if len(h.methods) == 0 {
		return true
	}
	_, f := h.methods[m]
	return f
}

func (h *prefixHandler) Name() string {
	return h.name
}

func newPrKeyHandler(name string, cfg *viper.Viper, ms *mongodb.SessionProvider) (HandlerWrap, error) {
	res := &prefixHandler{}
	err := initPrefixes(name, cfg, res)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't init prefix for %s", name)
	}
	res.h, err = newKeyHandler(name, cfg, ms)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init handler")
	}
	return res, nil
}

func newKeyHandler(name string, cfg *viper.Viper, ms *mongodb.SessionProvider) (http.Handler, error) {
	if cfg.GetString(name+".backend") == "" {
		return nil, errors.New("No backend")
	}
	goapp.Log.Infof("Backend: %s", cfg.GetString(name+".backend"))
	proxyURL := cfg.GetString(name + ".backend")
	url, err := utils.ParseURL(proxyURL)
	if err != nil {
		return nil, errors.Wrap(err, "Wrong backend")
	}

	dbProvider, err := mongodb.NewDBProvider(ms, cfg.GetString(name+".db"))
	if err != nil {
		return nil, errors.Wrap(err, "No db")
	}
	keysValidator, err := mongodb.NewKeyValidator(dbProvider)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init validator")
	}

	h := handler.FillOutHeader(handler.Proxy(url))
	h = handler.FillHeader(handler.FillKeyHeader(h))
	h, err = addCleanHeader(h, cfg.GetString(name+".cleanHeaderPrefix"))
	if err != nil {
		return nil, errors.Wrap(err, "can't init clean header")
	}

	stripURL := cfg.GetString(name + ".stripPrefix")
	if stripURL != "" {
		h = handler.StripPrefix(h, stripURL)
		goapp.Log.Infof("Strip prefix: %s", stripURL)
	}
	ls, err := mongodb.NewLogSaver(dbProvider)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init log saver")
	}
	h = handler.LogDB(h, ls)
	hKey := handler.KeyValid(h, keysValidator)
	h = handler.KeyExtract(hKey)

	return h, nil
}

func addCleanHeader(h http.Handler, headerPrefix string) (http.Handler, error) {
	res := h
	if headerPrefix != "" {
		var err error
		res, err = handler.CleanHeader(h, headerPrefix)
		if err != nil {
			return nil, errors.Wrap(err, "can't init clean header")
		}
	}
	return res, nil
}

func initMethods(str string) map[string]bool {
	res := make(map[string]bool)
	for _, s := range strings.Split(str, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			res[s] = true
		}
	}
	return res
}
