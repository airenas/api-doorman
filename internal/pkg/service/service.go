package service

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"

	"github.com/pkg/errors"
)

type (
	//IPManager manages IP in DB
	IPManager interface {
		CheckCreate(string, float64) error
	}

	//HandlerWrap for check if handler valid
	HandlerWrap interface {
		Valid(r *http.Request) bool
		Handler() http.Handler
		Info() string
		Name() string
	}

	//Data is service operation data
	Data struct {
		Port     int
		Handlers []HandlerWrap
	}
)

type mainHandler struct {
	data *Data
}

//StartWebServer starts the HTTP service and listens for the requests
func StartWebServer(data *Data) error {
	goapp.Log.Infof("Starting HTTP service at %d", data.Port)
	h, err := newMainHandler(data)
	if err != nil {
		return errors.Wrap(err, "Can't init handlers")
	}

	http.Handle("/", h)
	portStr := strconv.Itoa(data.Port)

	log(getInfo(data.Handlers))

	err = http.ListenAndServe(":"+portStr, nil)

	if err != nil {
		return errors.Wrap(err, "Can't start HTTP listener at port "+portStr)
	}
	return nil
}

func log(info string) {
	for _, s := range strings.Split(info, "\n") {
		goapp.Log.Info(s)
	}
}

func newMainHandler(data *Data) (http.Handler, error) {
	res := &mainHandler{}
	if len(data.Handlers) == 0 {
		return nil, errors.New("No handlers")
	}
	res.data = data
	return res, nil
}

func (h *mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, hi := range h.data.Handlers {
		if hi.Valid(r) {
			goapp.Log.Info("Handling with " + hi.Name())
			hi.Handler().ServeHTTP(w, r)
			return
		}
	}
	goapp.Log.Error("No handler for " + r.URL.Path)
	//serve not found
	http.NotFound(w, r)
}

func getInfo(handlers []HandlerWrap) string {
	sb := strings.Builder{}
	for _, h := range handlers {
		sb.WriteString(h.Info())
		sb.WriteString("\n")
	}
	return sb.String()
}
