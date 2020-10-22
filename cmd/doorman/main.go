package main

import (
	"log"

	"github.com/airenas/api-doorman/internal/pkg/service"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kelseyhightower/envconfig"
)

func main() {
	cfg := defaultConfig()
	envconfig.Process("", cfg)
	setLogLevel(cfg)

	data := service.Data{}
	data.Config = cfg

	err := service.StartWebServer(&data)
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
