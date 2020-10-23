package service

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type key int

const (
	cKey key = iota
	// ...
)

type (
	//Config is a struct to contain all the needed configuration for our Service
	Config struct {
		Port       int    `envconfig:"HTTP_PORT"`
		DebugLevel string `envconfig:"DEBUG_LEVEL"`
	}

	//Data is service operation data
	Data struct {
		Config *Config
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

func middlewareOne(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys, ok := r.URL.Query()["key"]

		if !ok || len(keys[0]) < 1 {
			log.Println("Url Param 'key' is missing")
			w.WriteHeader(401)
			return
		}

		// Query()["key"] will return an array of items,
		// we only want the single item.
		key := keys[0]

		log.Println("Url Param 'key' is: " + string(key))
		ctx := context.WithValue(r.Context(), cKey, key)

		log.Println("Executing middlewareOne")
		next.ServeHTTP(w, r.WithContext(ctx))
		log.Println("Executing middlewareOne again")
	})
}

func middlewareTwo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Executing middlewareTwo")
		key, _ := r.Context().Value(cKey).(string)
		log.Println("Url Param m2 'key' is: " + string(key))
		ip := getIP(r)
		log.Println("Request IP is: " + ip)

		if r.URL.Path == "/foo" {
			return
		}

		next.ServeHTTP(w, r)
		log.Println("Executing middlewareTwo again")
	})
}

func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return strings.Split(forwarded, ":")[0]
	}
	return r.RemoteAddr
}

//StartWebServer starts the HTTP service and listens for the requests
func StartWebServer(data *Data) error {
	logrus.Infof("Starting HTTP service at %d", data.Config.Port)
	http.Handle("/", middlewareOne(middlewareTwo(&mainHandler{})))
	portStr := strconv.Itoa(data.Config.Port)
	err := http.ListenAndServe(":"+portStr, nil)

	if err != nil {
		return errors.Wrap(err, "Can't start HTTP listener at port "+portStr)
	}
	return nil
}
