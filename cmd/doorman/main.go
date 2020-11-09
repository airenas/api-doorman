package main

import (
	"github.com/airenas/api-doorman/internal/pkg/audio"

	"github.com/airenas/api-doorman/internal/pkg/mongodb"
	"github.com/airenas/api-doorman/internal/pkg/service"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func main() {
	goapp.StartWithDefault()

	mongoSessionProvider, err := mongodb.NewSessionProvider(goapp.Config.GetString("mongo.url"))
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't init mongo provider"))
	}
	defer mongoSessionProvider.Close()

	data := service.Data{}
	data.Proxy = loadDataFromConfig(goapp.Sub(goapp.Config, "proxy"))
	data.Port = goapp.Config.GetInt("port")

	keysValidator, err := mongodb.NewKeyValidator(mongoSessionProvider)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't init saver"))
	}
	data.KeyValidator = keysValidator
	data.QuotaValidator = keysValidator

	saver, err := mongodb.NewLogSaver(mongoSessionProvider)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't init log saver"))
	}
	data.LogSaver = saver
	data.IPSaver, err = mongodb.NewIPSaver(mongoSessionProvider)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't init IP saver"))
	}
	dsURL := goapp.Config.GetString("proxy.quota.service")
	if dsURL != "" {
		data.DurationService, err = audio.NewDurationClient(dsURL)
		if err != nil {
			goapp.Log.Fatal(errors.Wrap(err, "Can't init Duration service"))
		}
		goapp.Log.Infof("Duration service: %s", dsURL)
	}

	err = service.StartWebServer(&data)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't start the service"))
	}
}

func loadDataFromConfig(cfg *viper.Viper) service.ProxyRoute {
	res := service.ProxyRoute{}
	res.BackendURL = cfg.GetString("backend")
	res.PrefixURL = cfg.GetString("prefixURL")
	res.Method = cfg.GetString("method")
	res.QuotaType = cfg.GetString("quota.type")
	res.QuotaField = cfg.GetString("quota.field")
	res.DefaultLimit = cfg.GetFloat64("quota.default")

	return res
}
