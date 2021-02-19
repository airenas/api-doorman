package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/mongodb"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type (
	// KeyCreator creates key
	KeyCreator interface {
		Create(string, *adminapi.Key) (*adminapi.Key, error)
	}

	// KeyUpdater creates key
	KeyUpdater interface {
		Update(string, string, map[string]interface{}) (*adminapi.Key, error)
	}

	// KeyRetriever gets keys list from db
	KeyRetriever interface {
		List(string) ([]*adminapi.Key, error)
	}

	// OneKeyRetriever retrieves one list from db
	OneKeyRetriever interface {
		Get(string, string) (*adminapi.Key, error)
	}

	// LogRetriever retrieves one list from db
	LogRetriever interface {
		Get(string, string) ([]*adminapi.Log, error)
	}

	// PrValidator validates if project is available
	PrValidator interface {
		Check(string) bool
	}

	//Data is service operation data
	Data struct {
		Port int

		KeySaver         KeyCreator
		KeyGetter        KeyRetriever
		OneKeyGetter     OneKeyRetriever
		LogGetter        LogRetriever
		OneKeyUpdater    KeyUpdater
		ProjectValidator PrValidator
	}
)

//StartWebServer starts the HTTP service and listens for the admin requests
func StartWebServer(data *Data) error {
	goapp.Log.Infof("Starting HTTP doorman admin service at %d", data.Port)
	r := NewRouter(data)
	portStr := strconv.Itoa(data.Port)

	w := goapp.Log.Writer()
	defer w.Close()
	l := log.New(w, "", 0)
	gracehttp.SetLogger(l)

	return gracehttp.Serve(&http.Server{Addr: ":" + portStr, Handler: r})
}

//NewRouter creates the router for HTTP service
func NewRouter(data *Data) *mux.Router {
	router := mux.NewRouter()
	router.Methods("POST").Path("/{project}/key").Handler(&keyAddHandler{data: data})
	router.Methods("GET").Path("/{project}/key-list").Handler(&keyListHandler{data: data})
	router.Methods("GET").Path("/{project}/key/{key}").Handler(&keyInfoHandler{data: data})
	router.Methods("PATCH").Path("/{project}/key/{key}").Handler(&keyUpdateHandler{data: data})
	return router
}

type keyAddHandler struct {
	data *Data
}

func (h *keyAddHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	goapp.Log.Infof("Request from %s", r.RemoteAddr)
	project := mux.Vars(r)["project"]
	if !validateProject(project, h.data.ProjectValidator, w) {
		return
	}
	decoder := json.NewDecoder(r.Body)
	var input adminapi.Key
	err := decoder.Decode(&input)
	if err != nil {
		http.Error(w, "Cannot decode input", http.StatusBadRequest)
		goapp.Log.Error("Cannot decode input" + err.Error())
		return
	}

	if input.Limit < 0.1 {
		http.Error(w, "No limit", http.StatusBadRequest)
		goapp.Log.Error("No input text")
		return
	}

	if input.ValidTo.Before(time.Now()) {
		http.Error(w, "Wrong valid to", http.StatusBadRequest)
		goapp.Log.Error("Wrong valid to")
		return
	}

	keyResp, err := h.data.KeySaver.Create(project, &input)

	if err != nil {
		if mongodb.IsDuplicate(err) {
			http.Error(w, "Duplicate key", http.StatusBadRequest)
		} else if errors.Is(err, adminapi.ErrWrongField) {
			http.Error(w, "Wrong field. "+err.Error(), http.StatusBadRequest)
			goapp.Log.Error("Wrong field. ", err)
			return
		} else {
			http.Error(w, "Service error", http.StatusInternalServerError)
		}
		goapp.Log.Error("Can't create key. ", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&keyResp)
	if err != nil {
		http.Error(w, "Can not prepare result", http.StatusInternalServerError)
		goapp.Log.Error(err)
	}
}

type keyListHandler struct {
	data *Data
}

func (h *keyListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	goapp.Log.Infof("Request list from %s", r.RemoteAddr)
	project := mux.Vars(r)["project"]
	if !validateProject(project, h.data.ProjectValidator, w) {
		return
	}
	keyResp, err := h.data.KeyGetter.List(project)

	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		goapp.Log.Error("Can't get keys. ", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&keyResp)
	if err != nil {
		http.Error(w, "Can not prepare result", http.StatusInternalServerError)
		goapp.Log.Error(err)
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
	goapp.Log.Infof("Request key from %s", r.RemoteAddr)
	key := mux.Vars(r)["key"]
	if key == "" {
		http.Error(w, "No Key", http.StatusBadRequest)
		goapp.Log.Errorf("No Key")
		return
	}
	project := mux.Vars(r)["project"]
	if !validateProject(project, h.data.ProjectValidator, w) {
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
	res.Key, err = h.data.OneKeyGetter.Get(project, key)
	if errors.Is(err, adminapi.ErrNoRecord) {
		http.Error(w, "Key not found", http.StatusBadRequest)
		goapp.Log.Error("Key not found.")
		return
	}
	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		goapp.Log.Error("Can't get key. ", err)
		return
	}
	if full {
		res.Logs, err = h.data.LogGetter.Get(project, key)
		if err != nil {
			http.Error(w, "Service error", http.StatusInternalServerError)
			goapp.Log.Error("Can't get logs. ", err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&res)
	if err != nil {
		http.Error(w, "Can not prepare result", http.StatusInternalServerError)
		goapp.Log.Error(err)
	}
}

type keyUpdateHandler struct {
	data *Data
}

func (h *keyUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	goapp.Log.Infof("Request from %s", r.RemoteAddr)

	key := mux.Vars(r)["key"]
	if key == "" {
		http.Error(w, "No Key", http.StatusBadRequest)
		goapp.Log.Errorf("No Key")
		return
	}
	project := mux.Vars(r)["project"]
	if !validateProject(project, h.data.ProjectValidator, w) {
		return
	}
	decoder := json.NewDecoder(r.Body)
	var input map[string]interface{}
	err := decoder.Decode(&input)
	if err != nil {
		http.Error(w, "Cannot decode input", http.StatusBadRequest)
		goapp.Log.Error("Cannot decode input" + err.Error())
		return
	}

	keyResp, err := h.data.OneKeyUpdater.Update(project, key, input)

	if errors.Is(err, adminapi.ErrNoRecord) {
		http.Error(w, "Key not found", http.StatusBadRequest)
		goapp.Log.Error("Key not found. ", err)
		return
	} else if errors.Is(err, adminapi.ErrWrongField) {
		http.Error(w, "Wrong field. "+err.Error(), http.StatusBadRequest)
		goapp.Log.Error("Wrong field. ", err)
		return
	} else if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		goapp.Log.Error("Can't update key. ", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&keyResp)
	if err != nil {
		http.Error(w, "Can not prepare result", http.StatusInternalServerError)
		goapp.Log.Error(err)
	}
}

func validateProject(project string, prV PrValidator, w http.ResponseWriter) bool {
	if project == "" {
		http.Error(w, "No Project", http.StatusBadRequest)
		goapp.Log.Errorf("No Project")
		return false
	}
	if !prV.Check(project) {
		http.Error(w, "Wrong project "+project, http.StatusBadRequest)
		goapp.Log.Errorf("Wrong project %s", project)
		return false
	}
	return true
}
