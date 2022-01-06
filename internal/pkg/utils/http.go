package utils

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func TakeJSONInput(c echo.Context, input interface{}) error {
	ctype := c.Request().Header.Get(echo.HeaderContentType)
	if !strings.HasPrefix(ctype, echo.MIMEApplicationJSON) {
		return echo.NewHTTPError(http.StatusBadRequest, "wrong content type, expected '"+echo.MIMEApplicationJSON+"'")
	}
	if err := c.Echo().JSONSerializer.Deserialize(c, input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "can't decode input")
	}
	return nil
}
