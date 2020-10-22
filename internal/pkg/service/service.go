package service

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Cannot read input", http.StatusInternalServerError)
		logrus.Errorln("Cannot read input", err)
		return
	}
	text := strings.TrimSpace(string(bodyBytes))
	logrus.Debugf("Input: %s", text)
	if text == "" {
		http.Error(w, "No input", http.StatusBadRequest)
		logrus.Errorln("No input", err)
		return
	}
}

//StartWebServer starts the HTTP service and listens for the requests
func StartWebServer(data *Data) error {
	logrus.Infof("Starting HTTP service at %d", data.Config.Port)
	r := NewRouter(data)
	http.Handle("/", r)
	portStr := strconv.Itoa(data.Config.Port)
	err := http.ListenAndServe(":"+portStr, nil)

	if err != nil {
		return errors.Wrap(err, "Can't start HTTP listener at port "+portStr)
	}
	return nil
}

//NewRouter creates the router for HTTP service
func NewRouter(data *Data) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.Methods("POST").Path("/tag").Handler(&mainHandler{data: data})
	return router
}
