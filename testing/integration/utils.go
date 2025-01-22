package integration

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	_admAuthKey = "olia"
)

func AddAdmAuth(request *http.Request) *http.Request {
	return AddAuth(request, _admAuthKey)
}

func AddAuth(req *http.Request, s string) *http.Request {
	req.Header.Add(echo.HeaderAuthorization, "Key "+s)
	return req
}

func Ptr[T any](in T) *T {
	return &in
}
