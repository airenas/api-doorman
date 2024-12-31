package admin

import (
	"fmt"
	slog "log"
	"net/http"
	"strconv"
	"strings"
	"time"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/integration/cms"
	"github.com/airenas/api-doorman/internal/pkg/mongodb"
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
	// KeyCreator creates key
	KeyCreator interface {
		Create(string, *adminapi.Key) (*adminapi.Key, error)
	}

	// KeyUpdater creates key
	KeyUpdater interface {
		Update(string, string, map[string]interface{}) (*adminapi.Key, error)
	}

	// UsageRestorer restores key usage by requestID
	UsageRestorer interface {
		RestoreUsage(project string, manual bool, request string, errorMsg string) error
	}

	// KeyRetriever gets keys list from db
	KeyRetriever interface {
		List(string) ([]*adminapi.Key, error)
	}

	// OneKeyRetriever retrieves one list from db
	OneKeyRetriever interface {
		Get(string, string) (*adminapi.Key, error)
	}

	// LogRetriever retrieves one list from db
	LogProvider interface {
		Get(string, string) ([]*adminapi.Log, error)
		List(string, time.Time) ([]*adminapi.Log, error)
		Delete(string, time.Time) (int, error)
	}

	// UsageReseter resets montly usage
	UsageReseter interface {
		Reset(ctx context.Context, project string, since time.Time, limit float64) error
	}

	// PrValidator validates if project is available
	PrValidator interface {
		Check(string) bool
		Projects() []string
	}

	//Data is service operation data
	Data struct {
		Port int

		KeySaver         KeyCreator
		KeyGetter        KeyRetriever
		OneKeyGetter     OneKeyRetriever
		LogProvider      LogProvider
		OneKeyUpdater    KeyUpdater
		UsageRestorer    UsageRestorer
		UsageReseter     UsageReseter
		ProjectValidator PrValidator

		CmsData *cms.Data
	}
)

// StartWebServer starts the HTTP service and listens for the admin requests
func StartWebServer(data *Data) error {
	log.Info().Msgf("Starting HTTP doorman admin service at %d", data.Port)

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

	e.GET("/live", live(data))
	e.GET("/:project/key-list", keyList(data))
	e.GET("/:project/key/:key", keyInfo(data))
	e.POST("/:project/key", keyAdd(data))
	e.PATCH("/:project/key/:key", keyUpdate(data))
	e.POST("/:project/restore/:requestID", restore(data))
	e.POST("/:project/reset", reset(data))
	e.GET("/:project/log", logList(data))
	e.DELETE("/:project/log", logDelete(data))

	cms.InitRoutes(e, data.CmsData)

	log.Info().Msg("Routes:")
	for _, r := range e.Routes() {
		log.Info().Msgf("  %s %s", r.Method, r.Path)
	}
	return e
}

func keyAdd(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		project := c.Param("project")
		if err := validateProject(project, data.ProjectValidator); err != nil {
			log.Error().Err(err).Send()
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		var input adminapi.Key
		if err := utils.TakeJSONInput(c, &input); err != nil {
			log.Error().Err(err).Send()
			return err
		}

		if input.Limit < 0.1 {
			log.Error().Msgf("no limit")
			return echo.NewHTTPError(http.StatusBadRequest, "no limit")
		}

		if input.ValidTo == nil || input.ValidTo.Before(time.Now()) {
			log.Error().Msgf("wrong valid to")
			return echo.NewHTTPError(http.StatusBadRequest, "wrong valid to")
		}

		keyResp, err := data.KeySaver.Create(project, &input)

		if err != nil {
			log.Error().Err(err).Msg("can't create key")
			if mongodb.IsDuplicate(err) {
				return echo.NewHTTPError(http.StatusBadRequest, "duplicate key")
			} else if errors.Is(err, adminapi.ErrWrongField) {
				return echo.NewHTTPError(http.StatusBadRequest, "wrong field. "+err.Error())
			} else {
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
		}
		return c.JSON(http.StatusOK, keyResp)
	}
}

func keyList(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		project := c.Param("project")
		if err := validateProject(project, data.ProjectValidator); err != nil {
			log.Error().Err(err).Send()
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		keyResp, err := data.KeyGetter.List(project)

		if err != nil {
			log.Error().Err(err).Send()
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, keyResp)
	}
}

func logList(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		project := c.Param("project")
		if err := validateProject(project, data.ProjectValidator); err != nil {
			log.Error().Err(err).Send()
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		to, err := utils.ParseDateParam(c.QueryParam("to"))
		if err != nil {
			return err
		}
		if to == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "no 'to' query param")
		}
		res, err := data.LogProvider.List(project, *to)

		if err != nil {
			log.Error().Err(err).Send()
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, res)
	}
}

func logDelete(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		project := c.Param("project")
		if err := validateProject(project, data.ProjectValidator); err != nil {
			log.Error().Err(err).Send()
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		to, err := utils.ParseDateParam(c.QueryParam("to"))
		if err != nil {
			return err
		}
		if to == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "no 'to' query param")
		}

		deleted, err := data.LogProvider.Delete(project, *to)
		if err != nil {
			log.Error().Err(err).Send()
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.JSONBlob(http.StatusOK, []byte(fmt.Sprintf(`{"deleted":%d}`, deleted)))
	}
}

func keyInfo(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
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
		res.Key, err = data.OneKeyGetter.Get(project, key)
		if errors.Is(err, adminapi.ErrNoRecord) {
			log.Error().Err(err).Send()
			return echo.NewHTTPError(http.StatusBadRequest, "key not found")
		}
		if err != nil {
			log.Error().Err(err).Send()
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		if c.QueryParam("full") == "1" {
			res.Logs, err = data.LogProvider.Get(project, key)
			if err != nil {
				log.Error().Err(err).Send()
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			if res.Key.ID != "" {
				logsByKeyID, err := data.LogProvider.Get(project, res.Key.ID)
				if err != nil {
					log.Error().Err(err).Send()
					return echo.NewHTTPError(http.StatusInternalServerError)
				}
				res.Logs = append(res.Logs, logsByKeyID...)
			}
		}
		return c.JSON(http.StatusOK, res)
	}
}

func keyUpdate(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
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
		input := make(map[string]interface{})
		if err := utils.TakeJSONInput(c, &input); err != nil {
			log.Error().Err(err).Send()
			return err
		}

		keyResp, err := data.OneKeyUpdater.Update(project, key, input)

		if errors.Is(err, adminapi.ErrNoRecord) {
			log.Error().Err(err).Msg("key not found")
			return echo.NewHTTPError(http.StatusBadRequest, "key not found")
		} else if errors.Is(err, adminapi.ErrWrongField) {
			log.Error().Err(err).Msg("wrong field")
			return echo.NewHTTPError(http.StatusBadRequest, "wrong field. "+err.Error())
		} else if err != nil {
			log.Error().Err(err).Msg("can't update key")
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, keyResp)
	}
}

type restoreReq struct {
	Error string `json:"error,omitempty"`
}

func restore(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
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

		err = data.UsageRestorer.RestoreUsage(project, manual, rID, input.Error)

		if errors.Is(err, adminapi.ErrNoRecord) {
			log.Error().Err(err).Msg("log not found")
			return echo.NewHTTPError(http.StatusBadRequest, "requestID not found")
		} else if errors.Is(err, adminapi.ErrLogRestored) {
			log.Error().Err(err).Msg("already restored")
			return echo.NewHTTPError(http.StatusConflict, "already restored")
		} else if err != nil {
			log.Error().Err(err).Msg("can't restore requestID usage")
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}
}

func reset(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		project := c.Param("project")
		if err := validateProject(project, data.ProjectValidator); err != nil {
			log.Error().Err(err).Send()
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		qStr := c.Param("quota")
		if qStr == "" {
			log.Error().Msgf("no quota")
			return echo.NewHTTPError(http.StatusBadRequest, "no quota")
		}
		quota, err := strconv.ParseFloat(qStr, 64)
		if err != nil {
			log.Error().Err(err).Send()
			return echo.NewHTTPError(http.StatusBadRequest, "wrong quota '%s'", qStr)
		}
		since, err := utils.ParseDateParam(c.Param("since"))
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
			log.Error().Err(err).Msg("reset usage")
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
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
