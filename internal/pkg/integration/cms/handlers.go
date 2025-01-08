package cms

import (
	"context"
	"net/http"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type (
	// Integrator wraps integratoin functionality
	Integrator interface {
		Create(ctx context.Context, in *api.CreateInput) (*api.Key, bool /*created*/, error)
		GetKey(ctx context.Context, id string) (*api.Key, error)
		AddCredits(ctx context.Context, id string, in *api.CreditsInput) (*api.Key, error)
		GetKeyID(ctx context.Context, id string) (*api.KeyID, error)
		Usage(ctx context.Context, id string, from *time.Time, to *time.Time, full bool) (*api.Usage, error)
		Update(ctx context.Context, id string, in map[string]interface{}) (*api.Key, error)
		Change(ctx context.Context, id string) (*api.Key, error)
		Changes(ctx context.Context, from *time.Time, projects []string) (*api.Changes, error)
	}

	// PrValidator validates if project is available
	PrValidator interface {
		Check(string) bool
		Projects() []string
	}

	//Data is main handler's data keeper
	Data struct {
		ProjectValidator PrValidator
		Integrator       Integrator
	}
)

// InitRoutes http routes for CMS integration
func InitRoutes(e *echo.Echo, data *Data) {
	e.POST("/key", keyCreate(data))
	e.POST("/keyID", keyGetID(data))
	e.GET("/key/:keyID", keyGet(data))
	e.PATCH("/key/:keyID", keyUpdate(data))
	e.PATCH("/key/:keyID/credits", keyAddCredits(data))
	e.POST("/key/:keyID/change", keyChange(data))
	e.GET("/key/:keyID/usage", keyUsage(data))
	e.GET("/keys/changes", keysChanges(data))
}

func keyCreate(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		var input api.CreateInput
		if err := utils.TakeJSONInput(c, &input); err != nil {
			log.Error().Err(err).Send()
			return err
		}
		if err := validateService(input.Service, data.ProjectValidator); err != nil {
			log.Error().Err(err).Send()
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		keyResp, created, err := data.Integrator.Create(c.Request().Context(), &input)

		if err != nil {
			return processError(err)
		}
		if created {
			return c.JSON(http.StatusCreated, keyResp)
		}
		return c.JSON(http.StatusConflict, keyResp)
	}
}

func keyGet(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		keyID := c.Param("keyID")
		if keyID == "" {
			log.Error().Msgf("no key ID")
			return echo.NewHTTPError(http.StatusBadRequest, "no key ID")
		}
		keyResp, err := data.Integrator.GetKey(c.Request().Context(), keyID)
		if err != nil {
			return processError(err)
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
			log.Error().Msgf("no key ID")
			return echo.NewHTTPError(http.StatusBadRequest, "no key ID")
		}
		input := make(map[string]interface{})
		if err := utils.TakeJSONInput(c, &input); err != nil {
			log.Error().Err(err).Send()
			return err
		}
		keyResp, err := data.Integrator.Update(c.Request().Context(), keyID, input)
		if err != nil {
			return processError(err)
		}
		return c.JSON(http.StatusOK, keyResp)
	}
}

func processError(err error) error {
	log.Error().Err(err).Send()
	var errF *utils.WrongFieldError
	if errors.As(err, &errF) {
		return echo.NewHTTPError(http.StatusBadRequest, errF.Error())
	}
	if errors.Is(err, utils.ErrNoRecord) {
		return echo.NewHTTPError(http.StatusBadRequest, "no record")
	}
	if errors.Is(err, utils.ErrDuplicate) {
		return echo.NewHTTPError(http.StatusBadRequest, "duplicate record")
	}
	if errors.Is(err, utils.ErrOperationExists) {
		return echo.NewHTTPError(http.StatusBadRequest, "duplicate operation")
	}
	return echo.NewHTTPError(http.StatusInternalServerError)
}

func keyAddCredits(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		keyID := c.Param("keyID")
		if keyID == "" {
			log.Error().Msgf("no key ID")
			return echo.NewHTTPError(http.StatusBadRequest, "no key ID")
		}
		var input api.CreditsInput
		if err := utils.TakeJSONInput(c, &input); err != nil {
			log.Error().Err(err).Send()
			return err
		}
		keyResp, err := data.Integrator.AddCredits(c.Request().Context(), keyID, &input)
		if err != nil {
			return processError(err)
		}
		return c.JSON(http.StatusOK, keyResp)
	}
}

func keyChange(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		keyID := c.Param("keyID")
		if keyID == "" {
			log.Error().Msg("no key ID")
			return echo.NewHTTPError(http.StatusBadRequest, "no key ID")
		}

		keyResp, err := data.Integrator.Change(c.Request().Context(), keyID)
		if err != nil {
			return processError(err)
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
			log.Error().Err(err).Send()
			return err
		}
		if input.Key == "" {
			log.Error().Msg("no key")
			return echo.NewHTTPError(http.StatusBadRequest, "no key")
		}
		keyResp, err := data.Integrator.GetKeyID(c.Request().Context(), input.Key)
		if err != nil {
			return processError(err)
		}

		return c.JSON(http.StatusOK, keyResp)
	}
}

func keyUsage(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		keyID := c.Param("keyID")
		if keyID == "" {
			log.Error().Msgf("no key ID")
			return echo.NewHTTPError(http.StatusBadRequest, "no key ID")
		}
		from, err := utils.ParseDateParam(c.QueryParam("from"))
		if err != nil {
			return err
		}
		to, err := utils.ParseDateParam(c.QueryParam("to"))
		if err != nil {
			return err
		}
		usageResp, err := data.Integrator.Usage(c.Request().Context(), keyID, from, to, c.QueryParam("full") == "1")
		if err != nil {
			return processError(err)
		}

		return c.JSON(http.StatusOK, usageResp)
	}
}

func keysChanges(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		from, err := utils.ParseDateParam(c.QueryParam("from"))
		if err != nil {
			return err
		}
		changesResp, err := data.Integrator.Changes(c.Request().Context(), from, data.ProjectValidator.Projects())
		if err != nil {
			return processError(err)
		}

		return c.JSON(http.StatusOK, changesResp)
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
