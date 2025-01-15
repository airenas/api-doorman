package utils

import (
	"net/http"
	"strings"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/model"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// TakeJSONInput extracts request body to 'input' as json object, checks for correct mime type
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

// NewTransport creates new roundtriper with same MaxIdleConnsPerHost
// ready to be used for access to only one host
func NewTransport() http.RoundTripper {
	res := http.DefaultTransport.(*http.Transport).Clone()
	res.MaxIdleConns = 100
	res.MaxConnsPerHost = 100
	res.MaxIdleConnsPerHost = 100
	res.IdleConnTimeout = time.Second * 90
	return res
}

func RunWithUser(c echo.Context, execFunc func(echo.Context, *model.User) error) error {
	defer goapp.Estimate("Service method: " + c.Path())()

	user, err := getUser(c)
	if err != nil {
		return err
	}
	return execFunc(c, user)
}

func getUser(c echo.Context) (*model.User, error) {
	r := c.Request()
	user, ok := r.Context().Value(model.CtxUser).(*model.User)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized)
	}
	if user.Disabled {
		return nil, echo.NewHTTPError(http.StatusUnauthorized)
	}
	return user, nil
}

func ProcessError(err error) error {
	log.Error().Err(err).Send()
	var errF *model.WrongFieldError
	if errors.As(err, &errF) {
		return echo.NewHTTPError(http.StatusBadRequest, errF.Error())
	}
	var errNA *model.NoAccessError
	if errors.As(err, &errNA) {
		return echo.NewHTTPError(http.StatusForbidden, errNA.Error())
	}
	if errors.Is(err, model.ErrNoRecord) {
		return echo.NewHTTPError(http.StatusBadRequest, "no record")
	}
	if errors.Is(err, model.ErrDuplicate) {
		return echo.NewHTTPError(http.StatusBadRequest, "duplicate record")
	}
	if errors.Is(err, model.ErrOperationExists) {
		return echo.NewHTTPError(http.StatusConflict, "duplicate operation")
	}
	if errors.Is(err, model.ErrOperationDiffers) {
		return echo.NewHTTPError(http.StatusBadRequest, "operation is not the same")
	}
	if errors.Is(err, model.ErrUnauthorized) {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	if errors.Is(err, model.ErrNoAccess) {
		return echo.NewHTTPError(http.StatusForbidden)
	}
	if errors.Is(err, model.ErrLogRestored) {
		return echo.NewHTTPError(http.StatusConflict, "already restored")
	}
	return echo.NewHTTPError(http.StatusInternalServerError)
}
