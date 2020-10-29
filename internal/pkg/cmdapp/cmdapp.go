package cmdapp

import (
	"os"
	"path/filepath"

	"github.com/heirko/go-contrib/logrusHelper"
	"github.com/heralight/logrus_mate"
	"github.com/pkg/errors"
)

//InitConfig tries to load config.yml from exe's dir
func InitConfig(configFile string) error {
	InitEnv(Config)

	failOnNoFail := false
	if configFile != "" {
		// Use config file from the flag.
		Config.SetConfigFile(configFile)
		failOnNoFail = true
	} else {
		// Find home directory.
		ex, err := os.Executable()
		if err != nil {
			return errors.Wrap(err, "Can't get the app directory")
		}
		Config.AddConfigPath(filepath.Dir(ex))
		Config.SetConfigName("config")
	}

	if err := Config.ReadInConfig(); err != nil {
		Log.Warn("Can't read config:", err)
		if failOnNoFail {
			return errors.Wrap(err, "Can't read config:")
		}
	}
	initLog()
	Log.Info("Config loaded from: ", Config.ConfigFileUsed())
	return nil
}

func initLog() {
	initDefaultLogConfig()
	c := logrusHelper.UnmarshalConfiguration(Config.Sub("logger"))
	initLogFromEnv(&c)
	err := logrusHelper.SetConfig(Log, c)
	if err != nil {
		Log.Error("Can't init log ", err)
	}
}

//initLogFromEnv tries to set level from environment
func initLogFromEnv(c *logrus_mate.LoggerConfig) {
	ll := Config.GetString("logger.level")
	if ll != "" {
		c.Level = ll
	}
}

func initDefaultLogConfig() {
	defaultLogConfig := map[string]interface{}{
		"level":                              "info",
		"formatter.name":                     "text",
		"formatter.options.full_timestamp":   true,
		"formatter.options.timestamp_format": "2006-01-02T15:04:05.000",
	}
	Config.SetDefault("logger", defaultLogConfig)
}
