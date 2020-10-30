package utils

import (
	"errors"
	"net/url"
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
