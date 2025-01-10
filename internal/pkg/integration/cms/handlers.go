package cms

import (
	"context"
	"net/http"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/model"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type (
	// Integrator wraps integratoin functionality
	Integrator interface {
		Create(ctx context.Context, user *model.User, in *api.CreateInput) (*api.Key, bool /*created*/, error)
		GetKey(ctx context.Context, user *model.User, id string) (*api.Key, error)
		AddCredits(ctx context.Context, user *model.User, id string, in *api.CreditsInput) (*api.Key, error)
		GetKeyID(ctx context.Context, user *model.User, id string) (*api.KeyID, error)
		Usage(ctx context.Context, user *model.User, id string, from *time.Time, to *time.Time, full bool) (*api.Usage, error)
		Update(ctx context.Context, user *model.User, id string, in *api.UpdateInput) (*api.Key, error)
		Change(ctx context.Context, user *model.User, id string) (*api.Key, error)
		Changes(ctx context.Context, user *model.User, from *time.Time, projects []string) (*api.Changes, error)
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
		return utils.RunWithUser(c, func(ctx echo.Context, u *model.User) error {
			var input api.CreateInput
			if err := utils.TakeJSONInput(c, &input); err != nil {
				log.Error().Err(err).Send()
				return err
			}
			if err := validateService(input.Service, data.ProjectValidator); err != nil {
				log.Error().Err(err).Send()
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}

			keyResp, created, err := data.Integrator.Create(c.Request().Context(), u, &input)

			if err != nil {
				return utils.ProcessError(err)
			}
			if created {
				return c.JSON(http.StatusCreated, keyResp)
			}
			return c.JSON(http.StatusConflict, keyResp)
		})
	}
}

func keyGet(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return utils.RunWithUser(c, func(ctx echo.Context, u *model.User) error {
			keyID := c.Param("keyID")
			if keyID == "" {
				log.Error().Msgf("no key ID")
				return echo.NewHTTPError(http.StatusBadRequest, "no key ID")
			}
			keyResp, err := data.Integrator.GetKey(c.Request().Context(), u, keyID)
			if err != nil {
				return utils.ProcessError(err)
			}
			if c.QueryParam("returnKey") != "1" {
				keyResp.Key = ""
			}
			return c.JSON(http.StatusOK, keyResp)
		})
	}
}

func keyUpdate(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return utils.RunWithUser(c, func(ctx echo.Context, u *model.User) error {
			keyID := c.Param("keyID")
			if keyID == "" {
				log.Error().Msgf("no key ID")
				return echo.NewHTTPError(http.StatusBadRequest, "no key ID")
			}
			var input api.UpdateInput
			if err := utils.TakeJSONInput(c, &input); err != nil {
				log.Error().Err(err).Send()
				return err
			}
			keyResp, err := data.Integrator.Update(c.Request().Context(), u, keyID, &input)
			if err != nil {
				return utils.ProcessError(err)
			}
			return c.JSON(http.StatusOK, keyResp)
		})
	}
}

func keyAddCredits(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return utils.RunWithUser(c, func(ctx echo.Context, u *model.User) error {
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
			keyResp, err := data.Integrator.AddCredits(c.Request().Context(), u, keyID, &input)
			if err != nil {
				return utils.ProcessError(err)
			}
			return c.JSON(http.StatusOK, keyResp)
		})
	}
}

func keyChange(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return utils.RunWithUser(c, func(ctx echo.Context, u *model.User) error {
			keyID := c.Param("keyID")
			if keyID == "" {
				log.Error().Msg("no key ID")
				return echo.NewHTTPError(http.StatusBadRequest, "no key ID")
			}

			keyResp, err := data.Integrator.Change(c.Request().Context(), u, keyID)
			if err != nil {
				return utils.ProcessError(err)
			}

			return c.JSON(http.StatusOK, keyResp)
		})
	}
}

type keyByIDInput struct {
	Key string `json:"key"`
}

func keyGetID(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return utils.RunWithUser(c, func(ctx echo.Context, u *model.User) error {
			var input keyByIDInput
			if err := utils.TakeJSONInput(c, &input); err != nil {
				log.Error().Err(err).Send()
				return err
			}
			if input.Key == "" {
				log.Error().Msg("no key")
				return echo.NewHTTPError(http.StatusBadRequest, "no key")
			}
			keyResp, err := data.Integrator.GetKeyID(c.Request().Context(), u, input.Key)
			if err != nil {
				return utils.ProcessError(err)
			}

			return c.JSON(http.StatusOK, keyResp)
		})
	}
}

func keyUsage(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return utils.RunWithUser(c, func(ctx echo.Context, u *model.User) error {
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
			usageResp, err := data.Integrator.Usage(c.Request().Context(), u, keyID, from, to, c.QueryParam("full") == "1")
			if err != nil {
				return utils.ProcessError(err)
			}

			return c.JSON(http.StatusOK, usageResp)
		})
	}
}

func keysChanges(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return utils.RunWithUser(c, func(ctx echo.Context, u *model.User) error {
			from, err := utils.ParseDateParam(c.QueryParam("from"))
			if err != nil {
				return err
			}
			changesResp, err := data.Integrator.Changes(c.Request().Context(), u, from, data.ProjectValidator.Projects())
			if err != nil {
				return utils.ProcessError(err)
			}

			return c.JSON(http.StatusOK, changesResp)
		})
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
