package main

import (
	"log"

	"github.com/airenas/api-doorman/internal/pkg/admin"
	"github.com/airenas/api-doorman/internal/pkg/mongodb"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
)

func main() {
	goapp.StartWithDefault()

	mongoSessionProvider, err := mongodb.NewSessionProvider(goapp.Config.GetString("mongo.url"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "Can't init mongo provider"))
	}
	defer mongoSessionProvider.Close()

	data := admin.Data{}
	data.Port = goapp.Config.GetInt("port")
	keysManager, err := mongodb.NewKeySaver(mongoSessionProvider, goapp.Config.GetInt("keySize"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "Can't init saver"))
	}
	data.KeyGetter, data.KeySaver, data.OneKeyUpdater = keysManager, keysManager, keysManager
	data.OneKeyGetter = keysManager

	logManager, err := mongodb.NewLogSaver(mongoSessionProvider)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Can't init log saver"))
	}
	data.LogGetter = logManager

	err = admin.StartWebServer(&data)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Can't start the service"))
	}
}
