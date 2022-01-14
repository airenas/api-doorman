package service

import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/facebookgo/grace/gracehttp"

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
		Priority() int
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
		return errors.Wrap(err, "can't init handlers")
	}

	portStr := strconv.Itoa(data.Port)

	logHandlers(getInfo(data.Handlers))

	w := goapp.Log.Writer()
	defer w.Close()
	gracehttp.SetLogger(log.New(w, "", 0))

	return gracehttp.Serve(&http.Server{
		Addr:        ":" + portStr,
		IdleTimeout: 10 * time.Minute, ReadHeaderTimeout: 20 * time.Second,
		ReadTimeout: 8 * time.Minute, WriteTimeout: 15 * time.Minute,
		Handler: h,
	})
}

func logHandlers(info string) {
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
	sort.Slice(res.data.Handlers, func(i, j int) bool { return res.data.Handlers[i].Priority() > res.data.Handlers[j].Priority() })
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
	goapp.Log.Error("no handler for " + r.URL.Path)
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
