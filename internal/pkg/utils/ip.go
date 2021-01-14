package utils

import (
	"net"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

//IPExtractor extracts IP from request
type IPExtractor interface {
	Get(*http.Request) string
}

//DefaultIPExtractor default implementation
var DefaultIPExtractor IPExtractor = &firstForwardFor{}

//ExtractIP default method to return ip from request
func ExtractIP(r *http.Request) string {
	if DefaultIPExtractor == nil {
		DefaultIPExtractor = &firstForwardFor{}
	}
	return DefaultIPExtractor.Get(r)
}

//NewIPExtractor creates new extractor based on type
func NewIPExtractor(ipType string) (IPExtractor, error) {
	if ipType == "" {
		return nil, errors.New("No ip extractor type ")
	}
	if ipType == "lastForwardFor" {
		return &lastForwardFor{}, nil
	}
	if ipType == "firstForwardFor" {
		return &firstForwardFor{}, nil
	}
	return nil, errors.Errorf("Unknown ip extractor type '%s'", ipType)
}

//GetIPHeader returns ip header
func GetIPHeader(r *http.Request) string {
	return r.Header.Get(ipHeader)
}

type lastForwardFor struct {
}

type firstForwardFor struct {
}

const ipHeader = "X-FORWARDED-FOR"

func (e *lastForwardFor) Get(r *http.Request) string {
	return getIP(r, getLast)
}

func (e *firstForwardFor) Get(r *http.Request) string {
	return getIP(r, getFirst)
}

func getIP(r *http.Request, f func(string) string) string {
	forwarded := r.Header.Get(ipHeader)
	if forwarded != "" {
		return trimPort(f(forwarded))
	}
	return trimPort(r.RemoteAddr)
}

func trimPort(s string) string {
	return strings.TrimSpace(strings.Split(s, ":")[0])
}

func getFirst(s string) string {
	return strings.Split(s, ",")[0]
}

func getLast(s string) string {
	strs := strings.Split(s, ",")
	return strs[len(strs)-1]
}

//ValidateIPsCIDR check if IPs are valid comma separated CIDR format
func ValidateIPsCIDR(ip string) error {
	for _, s := range strings.Split(ip, ",") {
		_, _, err := net.ParseCIDR(s)
		if err != nil {
			return errors.Wrapf(err, "Wrong IP: %s", s)
		}
	}
	return nil
}
