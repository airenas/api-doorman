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
	// KeyCreator creates key
	KeyCreator interface {
		Create(*adminapi.Key) (*adminapi.Key, error)
	}

	// KeyUpdater creates key
	KeyUpdater interface {
		Update(string, map[string]interface{}) (*adminapi.Key, error)
	}

	// KeyRetriever gets keys list from db
	KeyRetriever interface {
		List() ([]*adminapi.Key, error)
	}

	// OneKeyRetriever retrieves one list from db
	OneKeyRetriever interface {
		Get(key string) (*adminapi.Key, error)
	}

	// LogRetriever retrieves one list from db
	LogRetriever interface {
		Get(key string) ([]*adminapi.Log, error)
	}

	//Data is service operation data
	Data struct {
		Port int

		KeySaver      KeyCreator
		KeyGetter     KeyRetriever
		OneKeyGetter  OneKeyRetriever
		LogGetter     LogRetriever
		OneKeyUpdater KeyUpdater
	}
)

//StartWebServer starts the HTTP service and listens for the admin requests
func StartWebServer(data *Data) error {
	logrus.Infof("Starting HTTP doorman admin service at %d", data.Port)
	r := NewRouter(data)
	http.Handle("/", r)
	portStr := strconv.Itoa(data.Port)
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
	router.Methods("GET").Path("/key/{key}").Handler(&keyInfoHandler{data: data})
	router.Methods("PATCH").Path("/key/{key}").Handler(&keyUpdateHandler{data: data})
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

type keyInfoHandler struct {
	data *Data
}

type keyInfoResp struct {
	Key  *adminapi.Key   `json:"key,omitempty"`
	Logs []*adminapi.Log `json:"logs,omitempty"`
}

func (h *keyInfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("Request key from %s", r.Host)
	key := mux.Vars(r)["key"]
	if key == "" {
		http.Error(w, "No Key", http.StatusBadRequest)
		logrus.Errorf("No Key")
		return
	}
	query := r.URL.Query()
	qf, pf := query["full"]
	full := false
	if pf && len(qf) > 0 && qf[0] == "1" {
		full = true
	}

	res := &keyInfoResp{}
	var err error
	res.Key, err = h.data.OneKeyGetter.Get(key)
	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		logrus.Error("Can't get key. ", err)
		return
	}
	if res.Key == nil {
		http.Error(w, "Key not found", http.StatusBadRequest)
		logrus.Error("Key not found.")
		return
	}
	if full {
		res.Logs, err = h.data.LogGetter.Get(key)
		if err != nil {
			http.Error(w, "Service error", http.StatusInternalServerError)
			logrus.Error("Can't get logs. ", err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&res)
	if err != nil {
		http.Error(w, "Can not prepare result", http.StatusInternalServerError)
		logrus.Error(err)
	}
}

type keyUpdateHandler struct {
	data *Data
}

func (h *keyUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("Request from %s", r.Host)

	key := mux.Vars(r)["key"]
	if key == "" {
		http.Error(w, "No Key", http.StatusBadRequest)
		logrus.Errorf("No Key")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var input map[string]interface{}
	err := decoder.Decode(&input)
	if err != nil {
		http.Error(w, "Cannot decode input", http.StatusBadRequest)
		logrus.Error("Cannot decode input" + err.Error())
		return
	}

	keyResp, err := h.data.OneKeyUpdater.Update(key, input)

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
