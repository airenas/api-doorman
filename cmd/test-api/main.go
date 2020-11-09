package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/gorilla/mux"

	"github.com/pkg/errors"
)

func main() {
	port := flag.Int("p", 8000, "Port")
	goapp.StartWithDefault()

	err := startWebServer(*port)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't start the service"))
	}
}

func startWebServer(port int) error {
	goapp.Log.Infof("Starting test service on %d", port)
	http.Handle("/", newRouter())
	portStr := strconv.Itoa(port)
	err := http.ListenAndServe(":"+portStr, nil)
	if err != nil {
		return errors.Wrap(err, "Can't start HTTP listener at port "+portStr)
	}
	return nil
}

func newRouter() *mux.Router {
	router := mux.NewRouter()
	router.Methods("POST").Path("/public").Handler(&testHandler{name: "public"})
	router.Methods("POST").Path("/private").Handler(&testHandler{name: "private"})
	return router
}

type testHandler struct {
	name string
	num  int32
}

type response struct {
	Num  int32  `json:"num"`
	Name string `json:"name"`
	From string `json:"from"`
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	goapp.Log.Infof("Request from %s", r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	result := response{}
	result.From = r.RemoteAddr
	result.Name = h.name
	result.Num = atomic.AddInt32(&h.num, 1)
	err := encoder.Encode(&result)
	if err != nil {
		http.Error(w, "Can not prepare result", http.StatusInternalServerError)
		goapp.Log.Error(err)
	}
}
