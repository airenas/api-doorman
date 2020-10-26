package service

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/airenas/api-doorman/internal/pkg/handler"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type (

	//Config is a struct to contain all the needed configuration for our Service
	Config struct {
		Port       int    `envconfig:"HTTP_PORT"`
		DebugLevel string `envconfig:"DEBUG_LEVEL"`
		MongoURL   string `envconfig:"MONGO_URL"`
	}

	//IPManager manages IP in DB
	IPManager interface {
		CheckCreate(string, float64) error
	}

	//Data is service operation data
	Data struct {
		Config         *Config
		KeyValidator   handler.KeyValidator
		QuotaValidator handler.QuotaValidator
		LogSaver       handler.DBSaver
		IPSaver        IPManager
	}
)

type mainHandler struct {
	data *Data
}

func (t *mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// url, _ := url.Parse("http://localhost:80/")
	url, _ := url.Parse("http://list.airenas.eu:6080/")
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the headers to allow for SSL redirection
	r.URL.Host = url.Host
	r.URL.Scheme = url.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = url.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ModifyResponse = rewriteBody
	proxy.ServeHTTP(w, r)
}

func rewriteBody(resp *http.Response) (err error) {
	log.Printf("Resp: %d", resp.StatusCode)
	resp.Header.Set("olia", "oooo")
	return nil
}

//StartWebServer starts the HTTP service and listens for the requests
func StartWebServer(data *Data) error {
	logrus.Infof("Starting HTTP service at %d", data.Config.Port)
	// http.Handle("/", handler.NewKeyExtract(handler.KeyValid(
	// 	handler.RequestAsQuota(
	// 		handler.QuotaValidate(
	// 			&mainHandler{}, data.QuotaValidator)), data.KeyValidator)))

	h := handler.Proxy("http://list.airenas.eu:6080/")
	h = handler.QuotaValidate(h, data.QuotaValidator)
	h = handler.TakeJSON(handler.JSONAsQuota(h), "text")
	h = handler.LogDB(h, data.LogSaver)
	hKey := handler.KeyValid(h, data.KeyValidator)
	hIP := handler.IPAsKey(hKey, getIPSaver(data))
	hKeyIP := handler.KeyValidOrIP(hKey, hIP)
	h = handler.KeyExtract(hKeyIP)

	http.Handle("/", h)
	portStr := strconv.Itoa(data.Config.Port)
	err := http.ListenAndServe(":"+portStr, nil)

	if err != nil {
		return errors.Wrap(err, "Can't start HTTP listener at port "+portStr)
	}
	return nil
}

func getIPSaver(data *Data) handler.IPSaver {
	res := &ipSaver{}
	res.saver = data.IPSaver
	res.limit = 100
	return res
}
