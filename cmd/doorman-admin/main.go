package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/admin"
	"github.com/airenas/api-doorman/internal/pkg/mongodb"
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

	data := admin.Data{}
	data.Config = cfg
	keysManager, err := mongodb.NewKeySaver(mongoSessionProvider)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Can't init saver"))
	}
	data.KeyGetter, data.KeySaver = keysManager, keysManager

	err = admin.StartWebServer(&data)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Can't start the service"))
	}
}

func defaultConfig() *admin.Config {
	res := admin.Config{}
	res.Port = 8001
	res.DebugLevel = logrus.InfoLevel.String()
	return &res
}

func setLogLevel(cfg *admin.Config) {
	logrus.SetFormatter(&logrus.TextFormatter{TimestampFormat: "2006-01-02 15:04:05", FullTimestamp: true})
	l, err := logrus.ParseLevel(cfg.DebugLevel)
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.SetLevel(l)
}
