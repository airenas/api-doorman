package utils

import (
	"errors"
	"net/url"

	"github.com/airenas/go-app/pkg/goapp"
)

//ParseURL parse url and checks is not empty
func ParseURL(str string) (*url.URL, error) {
	u, err := url.Parse(str)
	if err != nil {
		return nil, err
	}
	if u.Host == "" {
		return nil, errors.New("Can't parse url")
	}
	return u, nil
}

//HidePass removes pass from URL
func HidePass(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		goapp.Log.Warn("Can't parse url.")
		return ""
	}
	if u.User != nil {
		u.User = url.UserPassword(u.User.Username(), "----")
	}
	return u.String()
}
