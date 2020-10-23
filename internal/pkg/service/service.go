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

	//Data is service operation data
	Data struct {
		Config         *Config
		KeyValidator   handler.KeyValidator
		QuotaValidator handler.QuotaValidator
	}
)

type mainHandler struct {
	data *Data
}

func (t *mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url, _ := url.Parse("http://localhost:80/")
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
	http.Handle("/", handler.NewKeyExtract(handler.KeyValid(
		handler.RequestAsQuota(
			handler.QuotaValidate(
				&mainHandler{}, data.QuotaValidator)), data.KeyValidator)))
	portStr := strconv.Itoa(data.Config.Port)
	err := http.ListenAndServe(":"+portStr, nil)

	if err != nil {
		return errors.Wrap(err, "Can't start HTTP listener at port "+portStr)
	}
	return nil
}
