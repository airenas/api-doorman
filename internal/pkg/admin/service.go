package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/gorilla/mux"
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

	// KeyCreator creates key
	KeyCreator interface {
		Create(*adminapi.Key) (*adminapi.Key, error)
	}

	// KeyRetriever gets keys list from db
	KeyRetriever interface {
		List() ([]*adminapi.Key, error)
	}

	//Data is service operation data
	Data struct {
		Config    *Config
		KeySaver  KeyCreator
		KeyGetter KeyRetriever
	}
)

//StartWebServer starts the HTTP service and listens for the admin requests
func StartWebServer(data *Data) error {
	logrus.Infof("Starting HTTP doorman admin service at %d", data.Config.Port)
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
	router := mux.NewRouter()
	router.Methods("POST").Path("/key").Handler(&keyAddHandler{data: data})
	router.Methods("GET").Path("/key-list").Handler(&keyListHandler{data: data})
	return router
}

type keyAddHandler struct {
	data *Data
}

func (h *keyAddHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("Request from %s", r.Host)

	decoder := json.NewDecoder(r.Body)
	var input adminapi.Key
	err := decoder.Decode(&input)
	if err != nil {
		http.Error(w, "Cannot decode input", http.StatusBadRequest)
		logrus.Error("Cannot decode input" + err.Error())
		return
	}

	if input.Limit < 0.1 {
		http.Error(w, "No limit", http.StatusBadRequest)
		logrus.Error("No input text")
		return
	}

	if input.ValidTo.Before(time.Now()) {
		http.Error(w, "Wrong valid to", http.StatusBadRequest)
		logrus.Error("Wrong valid to")
		return
	}

	keyResp, err := h.data.KeySaver.Create(&input)

	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		logrus.Error("Can't create key. ", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&keyResp)
	if err != nil {
		http.Error(w, "Can not prepare result", http.StatusInternalServerError)
		logrus.Error(err)
	}
}

type keyListHandler struct {
	data *Data
}

func (h *keyListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("Request list from %s", r.Host)

	keyResp, err := h.data.KeyGetter.List()

	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		logrus.Error("Can't get keys. ", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&keyResp)
	if err != nil {
		http.Error(w, "Can not prepare result", http.StatusInternalServerError)
		logrus.Error(err)
	}
}
