package handler

import (
	"log"
	"net/http"

	"github.com/sirupsen/logrus"
)

//KeyValidator validator
type KeyValidator interface {
	IsValid(string) (bool, error)
}

type keyValid struct {
	next http.Handler
	kv   KeyValidator
}

//KeyValid creates handler
func KeyValid(next http.Handler, kv KeyValidator) http.Handler {
	res := &keyValid{}
	res.kv = kv
	res.next = next
	return res
}

func (h *keyValid) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key, _ := r.Context().Value(CtxKey).(string)
	log.Println("Url Param 'key' is: " + string(key))

	ok, err := h.kv.IsValid(key)
	if err != nil {
		http.Error(w, "Service error", http.StatusInternalServerError)
		logrus.Error("Can't check key. ", err)
		return
	}
	if !ok {
		http.Error(w, "Key is not valid", http.StatusUnauthorized)
		return
	}

	if h.next != nil {
		h.next.ServeHTTP(w, r)
	}
}
