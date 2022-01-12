package cms

import (
	"fmt"
	"net/http"
	"time"

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
		AddCredits(string, *api.CreditsInput) (*api.Key, error)
		GetKeyID(string) (*api.KeyID, error)
		Usage(string, *time.Time, *time.Time, bool) (*api.Usage, error)
		Update(string, map[string]interface{}) (*api.Key, error)
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
	e.POST("/keyID", keyGetID(data))
	e.GET("/key/:keyID", keyGet(data))
	e.PATCH("/key/:keyID", keyUpdate(data))
	e.PATCH("/key/:keyID/credits", keyAddCredits(data))
	e.GET("/key/:keyID/usage", keyUsage(data))
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

func keyUpdate(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		keyID := c.Param("keyID")
		if keyID == "" {
			goapp.Log.Error("no key ID")
			return echo.NewHTTPError(http.StatusBadRequest, "no key ID")
		}
		input := make(map[string]interface{})
		if err := utils.TakeJSONInput(c, &input); err != nil {
			goapp.Log.Error(err)
			return err
		}
		keyResp, err := data.Integrator.Update(keyID, input)

		if err != nil {
			goapp.Log.Error("can't update key. ", err)
			var errF *api.ErrField
			if errors.As(err, &errF) {
				return echo.NewHTTPError(http.StatusBadRequest, errF.Error())
			}
			if errors.Is(err, api.ErrNoRecord) {
				return echo.NewHTTPError(http.StatusBadRequest, "no record by ID "+keyID)
			}
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, keyResp)
	}
}

func keyAddCredits(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		keyID := c.Param("keyID")
		if keyID == "" {
			goapp.Log.Error("no key ID")
			return echo.NewHTTPError(http.StatusBadRequest, "no key ID")
		}
		var input api.CreditsInput
		if err := utils.TakeJSONInput(c, &input); err != nil {
			goapp.Log.Error(err)
			return err
		}
		keyResp, err := data.Integrator.AddCredits(keyID, &input)

		if err != nil {
			goapp.Log.Error("can't add credits. ", err)
			var errF *api.ErrField
			if errors.As(err, &errF) {
				return echo.NewHTTPError(http.StatusBadRequest, errF.Error())
			}
			if errors.Is(err, api.ErrNoRecord) {
				return echo.NewHTTPError(http.StatusBadRequest, "no record by key ID")
			}
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, keyResp)
	}
}

type keyByIDInput struct {
	Key string `json:"key"`
}

func keyGetID(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		var input keyByIDInput
		if err := utils.TakeJSONInput(c, &input); err != nil {
			goapp.Log.Error(err)
			return err
		}
		if input.Key == "" {
			goapp.Log.Error("no key")
			return echo.NewHTTPError(http.StatusBadRequest, "no key")
		}
		keyResp, err := data.Integrator.GetKeyID(input.Key)

		if err != nil {
			goapp.Log.Error("can't get key by ID. ", err)
			if errors.Is(err, api.ErrNoRecord) {
				return echo.NewHTTPError(http.StatusBadRequest, "no record by key ")
			}
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, keyResp)
	}
}

func keyUsage(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		keyID := c.Param("keyID")
		if keyID == "" {
			goapp.Log.Error("no key ID")
			return echo.NewHTTPError(http.StatusBadRequest, "no key ID")
		}
		from, err := parseDateParam(c.QueryParam("from"))
		if err != nil {
			return err
		}
		to, err := parseDateParam(c.QueryParam("to"))
		if err != nil {
			return err
		}
		usageResp, err := data.Integrator.Usage(keyID, from, to, c.QueryParam("full") == "1")

		if err != nil {
			goapp.Log.Error("can't get usage. ", err)
			if errors.Is(err, api.ErrNoRecord) {
				return echo.NewHTTPError(http.StatusBadRequest, "no record by key ID")
			}
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, usageResp)
	}
}

func parseDateParam(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	res, err := time.Parse(time.RFC3339, s)
	if err != nil {
		goapp.Log.Error(errors.Wrapf(err, "can't parse as date '%s'", s))
		return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("can't parse as date '%s'", s))
	}
	return &res, nil
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
