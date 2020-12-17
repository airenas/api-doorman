package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/airenas/api-doorman/internal/pkg/audio"
	"github.com/airenas/api-doorman/internal/pkg/handler"
	"github.com/airenas/api-doorman/internal/pkg/mongodb"
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
		return newQuotaHandler(name, cfg, ms)
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

type quotaHandler struct {
	prefix   string
	method   string
	proxyURL string
	name     string
	h        http.Handler
}

func newQuotaHandler(name string, cfg *viper.Viper, ms *mongodb.SessionProvider) (HandlerWrap, error) {
	res := &quotaHandler{}
	res.name = name
	if cfg.GetString(name+".backend") == "" {
		return nil, errors.New("No backend")
	}
	goapp.Log.Infof("Backend: %s", cfg.GetString(name+".backend"))
	res.proxyURL = cfg.GetString(name + ".backend")
	url, err := utils.ParseURL(res.proxyURL)
	if err != nil {
		return nil, errors.Wrap(err, "Wrong backend")
	}
	res.prefix = strings.ToLower(cfg.GetString(name + ".prefixURL"))
	res.method = cfg.GetString(name + ".method")
	goapp.Log.Infof("PrefixURL: %s", res.prefix)

	dbProvider, err := mongodb.NewDBProvider(ms, cfg.GetString(name+".db"))
	if err != nil {
		return nil, errors.Wrap(err, "No db")
	}
	keysValidator, err := mongodb.NewKeyValidator(dbProvider)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init validator")
	}

	h := handler.QuotaValidate(handler.Proxy(url), keysValidator)
	qt := cfg.GetString(name + ".quota.type")
	qf := cfg.GetString(name + ".quota.field")
	if qt == "json" {
		if qf == "" {
			return nil, errors.New("No field")
		}
		goapp.Log.Infof("Quota extract: %s(%s)", qt, qf)
		h = handler.TakeJSON(handler.JSONAsQuota(h), qf)
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
	} else {
		return nil, errors.Errorf("Unknown proxy quota type '%s'", qf)
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
	res.h = handler.KeyExtract(hKey)

	return res, nil
}

func (h *quotaHandler) Handler() http.Handler {
	return h.h
}

func (h *quotaHandler) Info() string {
	return fmt.Sprintf("%s handler %s to '%s', prefix: %s", h.name, h.method, h.proxyURL, h.prefix)
}

func (h *quotaHandler) Valid(r *http.Request) bool {
	path := r.URL.Path
	path = strings.ToLower(path)
	return strings.HasPrefix(path, h.prefix) && h.method == r.Method
}

func (h *quotaHandler) Name() string {
	return h.name
}
