package cms

import (
	"net/http"

	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/mongodb"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type (
	// KeyCreator creates key
	Integrator interface {
		Create(*api.CreateInput) (*api.Key, bool, error)
		GetKey(keyID string) (*api.Key, error)
	}

	// PrValidator validates if project is available
	PrValidator interface {
		Check(string) bool
	}

	Data struct {
		ProjectValidator PrValidator
		Integrator       Integrator
	}
)

func InitRoutes(e *echo.Echo, data *Data) {
	e.POST("/key", keyCreate(data))
	e.GET("/key/:keyID", keyGet(data))
}

func keyCreate(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		var input api.CreateInput
		if err := utils.TakeJSONInput(c, &input); err != nil {
			goapp.Log.Error(err)
			return err
		}
		if err := validateService(input.Service, data.ProjectValidator); err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		keyResp, created, err := data.Integrator.Create(&input)

		if err != nil {
			goapp.Log.Error("can't create key. ", err)
			if mongodb.IsDuplicate(err) {
				return echo.NewHTTPError(http.StatusBadRequest, "duplicate key")
			}
			var errF *api.ErrField
			if errors.As(err, &errF) {
				return echo.NewHTTPError(http.StatusBadRequest, errF.Error())
			}
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		if created {
			return c.JSON(http.StatusCreated, keyResp)
		}
		return c.JSON(http.StatusOK, keyResp)
	}
}

func keyGet(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		keyID := c.Param("keyID")
		if keyID == "" {
			goapp.Log.Error("no key ID")
			return echo.NewHTTPError(http.StatusBadRequest, "no key ID")
		}
		keyResp, err := data.Integrator.GetKey(keyID)

		if err != nil {
			goapp.Log.Error("can't get key. ", err)
			if errors.Is(err, api.ErrNoRecord) {
				return echo.NewHTTPError(http.StatusBadRequest, "no record by ID "+keyID)
			}
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		if c.QueryParam("returnKey") != "1" {
			keyResp.Key = ""
		}
		return c.JSON(http.StatusOK, keyResp)
	}
}

func validateService(project string, prV PrValidator) error {
	if project == "" {
		return errors.New("no service")
	}
	if !prV.Check(project) {
		return errors.Errorf("wrong service %s", project)
	}
	return nil
}
