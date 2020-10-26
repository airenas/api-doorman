package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/mongodb"
	"github.com/airenas/api-doorman/internal/pkg/service"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kelseyhightower/envconfig"
)

func main() {
	cfg := defaultConfig()
	envconfig.Process("", cfg)
	setLogLevel(cfg)
	rand.Seed(time.Now().UnixNano())

	mongoSessionProvider, err := mongodb.NewSessionProvider(cfg.MongoURL)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Can't init mongo provider"))
	}
	defer mongoSessionProvider.Close()

	data := service.Data{}
	data.Config = cfg

	keysValidator, err := mongodb.NewKeyValidator(mongoSessionProvider)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Can't init saver"))
	}
	data.KeyValidator = keysValidator
	data.QuotaValidator = keysValidator

	saver, err := mongodb.NewLogSaver(mongoSessionProvider)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Can't init log saver"))
	}
	data.LogSaver = saver
	data.IPSaver, err = mongodb.NewIPSaver(mongoSessionProvider)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Can't init IP saver"))
	}

	err = service.StartWebServer(&data)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Can't start the service"))
	}
}

func defaultConfig() *service.Config {
	res := service.Config{}
	res.Port = 8000
	res.DebugLevel = logrus.InfoLevel.String()
	return &res
}

func setLogLevel(cfg *service.Config) {
	logrus.SetFormatter(&logrus.TextFormatter{TimestampFormat: "2006-01-02 15:04:05", FullTimestamp: true})
	l, err := logrus.ParseLevel(cfg.DebugLevel)
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.SetLevel(l)
}
