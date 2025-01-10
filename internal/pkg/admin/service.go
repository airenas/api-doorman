package admin

import (
	slog "log"
	"net/http"
	"strconv"
	"strings"
	"time"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/integration/cms"
	"github.com/airenas/api-doorman/internal/pkg/model"
	"github.com/airenas/api-doorman/internal/pkg/model/permission"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

type (
	// UsageRestorer restores key usage by requestID
	UsageRestorer interface {
		RestoreUsage(ctx context.Context, project string, manual bool, request string, errorMsg string) error
	} // OneKeyRetriever retrieves one list from db
	OneKeyRetriever interface {
		Get(ctx context.Context, user *model.User, id string) (*adminapi.Key, error)
	}
	// LogRetriever retrieves one list from db
	LogProvider interface {
		GetLogs(ctx context.Context, user *model.User, keyID string) ([]*adminapi.Log, error)
		ListLogs(ctx context.Context, project string, to time.Time) ([]*adminapi.Log, error)
		DeleteLogs(ctx context.Context, project string, to time.Time) (int /* count of deleted items*/, error)
	}

	// UsageReseter resets montly usage
	UsageReseter interface {
		Reset(ctx context.Context, project string, since time.Time, limit float64) error
	}
	Auth interface {
		ValidateToken(ctx context.Context, token string) (model.User, error)
	}

	// PrValidator validates if project is available
	PrValidator interface {
		Check(string) bool
		Projects() []string
	}

	Hasher interface {
		HashKey(key string) string
	}

	//Data is service operation data
	Data struct {
		Port         int
		OneKeyGetter OneKeyRetriever
		LogProvider  LogProvider
		// logProvider       LogProvider
		UsageRestorer    UsageRestorer
		UsageReseter     UsageReseter
		ProjectValidator PrValidator
		Auth             echo.MiddlewareFunc
		Hasher           Hasher

		CmsData *cms.Data
	}
)

// StartWebServer starts the HTTP service and listens for the admin requests
func StartWebServer(data *Data) error {
	if data == nil {
		return errors.New("no data")
	}
	if data.Auth == nil {
		return errors.New("no auth")
	}
	if data.Hasher == nil {
		return errors.New("no hasher")
	}
	if data.OneKeyGetter == nil {
		return errors.New("no OneKeyGetter")
	}
	if data.LogProvider == nil {
		return errors.New("no LogProvider")
	}

	log.Info().Int("port", data.Port).Msg("Starting HTTP doorman admin service")

	e := initRoutes(data)

	portStr := strconv.Itoa(data.Port)
	e.Server.Addr = ":" + portStr
	e.Server.IdleTimeout = 3 * time.Minute
	e.Server.ReadHeaderTimeout = 10 * time.Second
	e.Server.ReadTimeout = 20 * time.Second
	e.Server.WriteTimeout = 60 * time.Second

	gracehttp.SetLogger(slog.New(goapp.Log, "", 0))

	return gracehttp.Serve(e.Server)
}

var promMdlw *prometheus.Prometheus

func init() {
	promMdlw = prometheus.NewPrometheus("doorman_admin", nil)
}

func initRoutes(data *Data) *echo.Echo {
	e := echo.New()
	promMdlw.Use(e)
	e.Use(data.Auth)

	e.GET("/live", live(data))
	e.POST("/hash", makeHash(data))
	e.GET("/:project/key/:key", keyInfo(data))
	e.POST("/:project/restore/:requestID", restore(data))
	e.POST("/:project/reset", reset(data))
	// e.GET("/:project/log", logList(data))
	// e.DELETE("/:project/log", logDelete(data))

	cms.InitRoutes(e, data.CmsData)

	log.Info().Msg("Routes:")
	for _, r := range e.Routes() {
		log.Info().Msgf("  %s %s", r.Method, r.Path)
	}
	return e
}

// func logList(data *Data) func(echo.Context) error {
// 	return func(c echo.Context) error {
// 		defer goapp.Estimate("Service method: " + c.Path())()
// 		project := c.Param("project")
// 		if err := validateProject(project, data.ProjectValidator); err != nil {
// 			log.Error().Err(err).Send()
// 			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 		}
// 		to, err := utils.ParseDateParam(c.QueryParam("to"))
// 		if err != nil {
// 			return err
// 		}
// 		if to == nil {
// 			return echo.NewHTTPError(http.StatusBadRequest, "no 'to' query param")
// 		}
// 		res, err := data.LogProvider.ListLogs(c.Request().Context(), project, *to)

// 		if err != nil {
// 			return utils.ProcessError(err)
// 		}

// 		return c.JSON(http.StatusOK, res)
// 	}
// }

// func logDelete(data *Data) func(echo.Context) error {
// 	return func(c echo.Context) error {
// 		defer goapp.Estimate("Service method: " + c.Path())()
// 		project := c.Param("project")
// 		if err := validateProject(project, data.ProjectValidator); err != nil {
// 			log.Error().Err(err).Send()
// 			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 		}
// 		to, err := utils.ParseDateParam(c.QueryParam("to"))
// 		if err != nil {
// 			return err
// 		}
// 		if to == nil {
// 			return echo.NewHTTPError(http.StatusBadRequest, "no 'to' query param")
// 		}

// 		deleted, err := data.LogProvider.DeleteLogs(c.Request().Context(), project, *to)
// 		if err != nil {
// 			return utils.ProcessError(err)
// 		}

// 		return c.JSONBlob(http.StatusOK, []byte(fmt.Sprintf(`{"deleted":%d}`, deleted)))
// 	}
// }

func keyInfo(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return utils.RunWithUser(c, func(e echo.Context, user *model.User) error {
			key := c.Param("key")
			if key == "" {
				log.Error().Msgf("no key")
				return echo.NewHTTPError(http.StatusBadRequest, "no key")
			}
			project := c.Param("project")
			if err := validateProject(project, data.ProjectValidator); err != nil {
				log.Error().Err(err).Send()
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}

			res := &adminapi.KeyInfoResp{}
			var err error
			res.Key, err = data.OneKeyGetter.Get(c.Request().Context(), user, key)
			if err != nil {
				return utils.ProcessError(err)
			}

			if c.QueryParam("full") == "1" {
				res.Logs, err = data.LogProvider.GetLogs(c.Request().Context(), user, key)
				if err != nil {
					return utils.ProcessError(err)
				}
			}
			return c.JSON(http.StatusOK, res)
		})
	}
}

func makeHash(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return utils.RunWithUser(c, func(e echo.Context, user *model.User) error {
			if !user.HasPermission(permission.Everything) {
				return echo.NewHTTPError(http.StatusForbidden)
			}
			var input adminapi.KeyIn
			if err := utils.TakeJSONInput(c, &input); err != nil {
				log.Error().Err(err).Send()
				return err
			}
			res := data.Hasher.HashKey(input.Key)
			return c.String(http.StatusOK, res)
		})
	}
}

type restoreReq struct {
	Error string `json:"error,omitempty"`
}

func restore(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return utils.RunWithUser(c, func(e echo.Context, user *model.User) error {
			if !user.HasPermission(permission.RestoreUsage) {
				return echo.NewHTTPError(http.StatusForbidden)
			}
			defer goapp.Estimate("Service method: " + c.Path())()
			rStr := c.Param("requestID")
			if rStr == "" {
				log.Error().Msgf("no requestID")
				return echo.NewHTTPError(http.StatusBadRequest, "no requestID")
			}
			rID, manual, err := parseRequestID(rStr)
			if err != nil {
				log.Error().Err(err).Send()
				return echo.NewHTTPError(http.StatusBadRequest, "wrong requestID format")
			}
			project := c.Param("project")
			if err := validateProject(project, data.ProjectValidator); err != nil {
				log.Error().Err(err).Send()
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			input := restoreReq{}
			if err := utils.TakeJSONInput(c, &input); err != nil {
				log.Error().Err(err).Send()
				return err
			}

			err = data.UsageRestorer.RestoreUsage(c.Request().Context(), project, manual, rID, input.Error)
			if err != nil {
				return utils.ProcessError(err)
			}
			return c.NoContent(http.StatusOK)
		})
	}
}

func reset(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return utils.RunWithUser(c, func(ctx echo.Context, u *model.User) error {
			if !u.HasPermission(permission.ResetMonthlyUsage) {
				return echo.NewHTTPError(http.StatusForbidden)
			}
			project := c.Param("project")
			if err := validateProject(project, data.ProjectValidator); err != nil {
				log.Error().Err(err).Send()
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			qStr := c.QueryParam("quota")
			if qStr == "" {
				log.Error().Msgf("no quota")
				return echo.NewHTTPError(http.StatusBadRequest, "no quota")
			}
			quota, err := strconv.ParseFloat(qStr, 64)
			if err != nil {
				log.Error().Err(err).Send()
				return echo.NewHTTPError(http.StatusBadRequest, "wrong quota '%s'", qStr)
			}
			since, err := utils.ParseDateParam(c.QueryParam("since"))
			if err != nil {
				log.Error().Err(err).Send()
				return echo.NewHTTPError(http.StatusBadRequest, "wrong since '%s'", qStr)
			}
			if since == nil {
				log.Error().Msgf("no since")
				return echo.NewHTTPError(http.StatusBadRequest, "no since")
			}

			err = data.UsageReseter.Reset(c.Request().Context(), project, *since, quota)

			if err != nil {
				return utils.ProcessError(err)
			}

			return c.NoContent(http.StatusOK)
		})
	}
}

func parseRequestID(rStr string) (string, bool, error) {
	strs := strings.Split(rStr, ":")
	if len(strs) != 2 {
		return "", false, errors.Errorf("wrong request format %s, wanted 'manual:rID'", rStr)
	}
	return strs[1], strs[0] == "m", nil
}

func validateProject(project string, prV PrValidator) error {
	if project == "" {
		return errors.New("no project")
	}
	if !prV.Check(project) {
		return errors.Errorf("wrong project %s", project)
	}
	return nil
}

func live(_data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, []byte(`{"service":"OK"}`))
	}
}
