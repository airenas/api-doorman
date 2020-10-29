package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/audio"

	"github.com/airenas/api-doorman/internal/pkg/cmdapp"
	"github.com/airenas/api-doorman/internal/pkg/mongodb"
	"github.com/airenas/api-doorman/internal/pkg/service"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func main() {
	cFile := flag.String("c", "", "Config yml file")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:[params] \n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	err := cmdapp.InitConfig(*cFile)
	if err != nil {
		cmdapp.Log.Fatal(errors.Wrap(err, "Can't init app"))
	}

	rand.Seed(time.Now().UnixNano())

	mongoSessionProvider, err := mongodb.NewSessionProvider(cmdapp.Config.GetString("mongo.url"))
	if err != nil {
		cmdapp.Log.Fatal(errors.Wrap(err, "Can't init mongo provider"))
	}
	defer mongoSessionProvider.Close()

	data := service.Data{}
	data.Proxy, err = loadDataFromConfig(cmdapp.Sub(cmdapp.Config, "proxy"))
	data.Port = cmdapp.Config.GetInt("port")

	keysValidator, err := mongodb.NewKeyValidator(mongoSessionProvider)
	if err != nil {
		cmdapp.Log.Fatal(errors.Wrap(err, "Can't init saver"))
	}
	data.KeyValidator = keysValidator
	data.QuotaValidator = keysValidator

	saver, err := mongodb.NewLogSaver(mongoSessionProvider)
	if err != nil {
		cmdapp.Log.Fatal(errors.Wrap(err, "Can't init log saver"))
	}
	data.LogSaver = saver
	data.IPSaver, err = mongodb.NewIPSaver(mongoSessionProvider)
	if err != nil {
		cmdapp.Log.Fatal(errors.Wrap(err, "Can't init IP saver"))
	}
	dsURL := cmdapp.Config.GetString("proxy.quota.service")
	if dsURL != "" {
		data.DurationService, err = audio.NewDurationClient(dsURL)
		if err != nil {
			cmdapp.Log.Fatal(errors.Wrap(err, "Can't init Duration service"))
		}
		cmdapp.Log.Infof("Duration service: %s", dsURL)
	}

	err = service.StartWebServer(&data)
	if err != nil {
		cmdapp.Log.Fatal(errors.Wrap(err, "Can't start the service"))
	}
}

func loadDataFromConfig(cfg *viper.Viper) (service.ProxyRoute, error) {
	res := service.ProxyRoute{}
	res.BackendURL = cfg.GetString("backend")
	res.PrefixURL = cfg.GetString("prefixURL")
	res.Method = cfg.GetString("method")
	res.QuotaType = cfg.GetString("quota.type")
	res.QuotaField = cfg.GetString("quota.field")
	res.DefaultLimit = cfg.GetFloat64("quota.default")

	return res, nil
}
