package service

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/airenas/api-doorman/internal/pkg/cmdapp"
	"github.com/airenas/api-doorman/internal/pkg/handler"
	"github.com/pkg/errors"
)

type (
	//IPManager manages IP in DB
	IPManager interface {
		CheckCreate(string, float64) error
	}

	//ProxyRoute keeps config for one proxy route
	ProxyRoute struct {
		BackendURL   string
		PrefixURL    string
		Method       string
		QuotaType    string
		QuotaField   string
		DefaultLimit float64
	}

	//Data is service operation data
	Data struct {
		Port            int
		KeyValidator    handler.KeyValidator
		QuotaValidator  handler.QuotaValidator
		DurationService handler.AudioLenGetter
		LogSaver        handler.DBSaver
		IPSaver         IPManager
		Proxy           ProxyRoute
	}
)

type mainHandler struct {
	data     *Data
	handlers []*hWrap
	def      http.Handler
}

type hWrap struct {
	prefix string
	method string
	h      http.Handler
}

func (h *mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	path = strings.ToLower(path)
	for _, hi := range h.handlers {
		if strings.HasPrefix(path, hi.prefix) && hi.method == r.Method {
			cmdapp.Log.Info("Serving " + hi.prefix)
			hi.h.ServeHTTP(w, r)
			return
		}
	}
	cmdapp.Log.Info("Serving default")
	h.def.ServeHTTP(w, r)
}

//StartWebServer starts the HTTP service and listens for the requests
func StartWebServer(data *Data) error {
	cmdapp.Log.Infof("Starting HTTP service at %d", data.Port)
	h, err := newMainHandler(data)
	if err != nil {
		return errors.Wrap(err, "Can't init handlers")
	}

	http.Handle("/", h)
	portStr := strconv.Itoa(data.Port)
	err = http.ListenAndServe(":"+portStr, nil)

	if err != nil {
		return errors.Wrap(err, "Can't start HTTP listener at port "+portStr)
	}
	return nil
}

func newMainHandler(data *Data) (http.Handler, error) {
	res := &mainHandler{}
	if data.Proxy.BackendURL == "" {
		return nil, errors.New("No backend")
	}
	cmdapp.Log.Infof("Backend: %s", data.Proxy.BackendURL)
	res.def = handler.Proxy(data.Proxy.BackendURL)
	if data.Proxy.PrefixURL == "" {
		return nil, errors.New("No prefix URL")
	}
	if data.Proxy.Method == "" {
		return nil, errors.New("No proxy method")
	}
	if data.Proxy.QuotaType == "" {
		return nil, errors.New("No proxy quota type")
	}
	hw := &hWrap{}
	res.handlers = append(res.handlers, hw)
	hw.prefix = strings.ToLower(data.Proxy.PrefixURL)
	hw.method = data.Proxy.Method
	cmdapp.Log.Infof("PrefixURL: %s", hw.prefix)

	h := handler.QuotaValidate(res.def, data.QuotaValidator)
	if data.Proxy.QuotaType == "json" {
		cmdapp.Log.Infof("Quota extract: %s(%s)", data.Proxy.QuotaType, data.Proxy.QuotaField)
		h = handler.TakeJSON(handler.JSONAsQuota(h), data.Proxy.QuotaField)
	} else if data.Proxy.QuotaType == "audioDuration" {
		if data.DurationService == nil {
			return nil, errors.New("No duration service initialized")
		}
		cmdapp.Log.Infof("Quota extract: %s(%s) using duration service", data.Proxy.QuotaType, data.Proxy.QuotaField)
		h = handler.AudioLenQuota(h, data.Proxy.QuotaField, data.DurationService)
	} else {
		return nil, errors.Errorf("Unknown proxy quota type '%s'", data.Proxy.QuotaType)
	}
	h = handler.LogDB(h, data.LogSaver)
	hKey := handler.KeyValid(h, data.KeyValidator)
	if data.Proxy.DefaultLimit > 0 {
		cmdapp.Log.Infof("Default IP quota: %.f", data.Proxy.DefaultLimit)
		hIP := handler.IPAsKey(hKey, newIPSaver(data))
		hKey = handler.KeyValidOrIP(hKey, hIP)
	}
	hw.h = handler.KeyExtract(hKey)

	return res, nil
}
