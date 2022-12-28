package admin

import (
	"log"
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
	LogRetriever interface {
		Get(string, string) ([]*adminapi.Log, error)
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
		LogGetter        LogRetriever
		OneKeyUpdater    KeyUpdater
		UsageRestorer    UsageRestorer
		ProjectValidator PrValidator

		CmsData *cms.Data
	}
)

//StartWebServer starts the HTTP service and listens for the admin requests
func StartWebServer(data *Data) error {
	goapp.Log.Infof("Starting HTTP doorman admin service at %d", data.Port)

	e := initRoutes(data)

	portStr := strconv.Itoa(data.Port)
	e.Server.Addr = ":" + portStr
	e.Server.IdleTimeout = 3 * time.Minute
	e.Server.ReadHeaderTimeout = 10 * time.Second
	e.Server.ReadTimeout = 20 * time.Second
	e.Server.WriteTimeout = 60 * time.Second

	w := goapp.Log.Writer()
	defer w.Close()
	gracehttp.SetLogger(log.New(w, "", 0))

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

	cms.InitRoutes(e, data.CmsData)

	goapp.Log.Info("Routes:")
	for _, r := range e.Routes() {
		goapp.Log.Infof("  %s %s", r.Method, r.Path)
	}
	return e
}

func keyAdd(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		project := c.Param("project")
		if err := validateProject(project, data.ProjectValidator); err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		var input adminapi.Key
		if err := utils.TakeJSONInput(c, &input); err != nil {
			goapp.Log.Error(err)
			return err
		}

		if input.Limit < 0.1 {
			goapp.Log.Error("no limit")
			return echo.NewHTTPError(http.StatusBadRequest, "no limit")
		}

		if input.ValidTo == nil || input.ValidTo.Before(time.Now()) {
			goapp.Log.Error("wrong valid to")
			return echo.NewHTTPError(http.StatusBadRequest, "wrong valid to")
		}

		keyResp, err := data.KeySaver.Create(project, &input)

		if err != nil {
			goapp.Log.Error("can't create key. ", err)
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
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		keyResp, err := data.KeyGetter.List(project)

		if err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, keyResp)
	}
}

func keyInfo(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: " + c.Path())()
		key := c.Param("key")
		if key == "" {
			goapp.Log.Error("no key")
			return echo.NewHTTPError(http.StatusBadRequest, "no key")
		}
		project := c.Param("project")
		if err := validateProject(project, data.ProjectValidator); err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		res := &adminapi.KeyInfoResp{}
		var err error
		res.Key, err = data.OneKeyGetter.Get(project, key)
		if errors.Is(err, adminapi.ErrNoRecord) {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusBadRequest, "key not found")
		}
		if err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		if c.QueryParam("full") == "1" {
			res.Logs, err = data.LogGetter.Get(project, key)
			if err != nil {
				goapp.Log.Error(err)
				return echo.NewHTTPError(http.StatusInternalServerError)
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
			goapp.Log.Error("no key")
			return echo.NewHTTPError(http.StatusBadRequest, "no key")
		}
		project := c.Param("project")
		if err := validateProject(project, data.ProjectValidator); err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		input := make(map[string]interface{})
		if err := utils.TakeJSONInput(c, &input); err != nil {
			goapp.Log.Error(err)
			return err
		}

		keyResp, err := data.OneKeyUpdater.Update(project, key, input)

		if errors.Is(err, adminapi.ErrNoRecord) {
			goapp.Log.Error("key not found. ", err)
			return echo.NewHTTPError(http.StatusBadRequest, "key not found")
		} else if errors.Is(err, adminapi.ErrWrongField) {
			goapp.Log.Error("wrong field. ", err)
			return echo.NewHTTPError(http.StatusBadRequest, "wrong field. "+err.Error())
		} else if err != nil {
			goapp.Log.Error("can't update key. ", err)
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
			goapp.Log.Error("no requestID")
			return echo.NewHTTPError(http.StatusBadRequest, "no requestID")
		}
		rID, manual, err := parseRequestID(rStr)
		if err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusBadRequest, "wrong requestID format")
		}
		project := c.Param("project")
		if err := validateProject(project, data.ProjectValidator); err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		input := restoreReq{}
		if err := utils.TakeJSONInput(c, &input); err != nil {
			goapp.Log.Error(err)
			return err
		}

		err = data.UsageRestorer.RestoreUsage(project, manual, rID, input.Error)

		if errors.Is(err, adminapi.ErrNoRecord) {
			goapp.Log.Error("log not found. ", err)
			return echo.NewHTTPError(http.StatusBadRequest, "requestID not found")
		} else if errors.Is(err, adminapi.ErrLogRestored) {
			goapp.Log.Error("already restored. ", err)
			return echo.NewHTTPError(http.StatusConflict, "already restored")
		} else if err != nil {
			goapp.Log.Error("can't restore requestID usage. ", err)
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

func live(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, []byte(`{"service":"OK"}`))
	}
}
