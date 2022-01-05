package main

import (
	"log"

	"github.com/airenas/api-doorman/internal/pkg/admin"
	"github.com/airenas/api-doorman/internal/pkg/mongodb"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/labstack/gommon/color"
	"github.com/pkg/errors"
)

func main() {
	goapp.StartWithDefault()

	mongoSessionProvider, err := mongodb.NewSessionProvider(goapp.Config.GetString("mongo.url"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't init mongo provider"))
	}
	defer mongoSessionProvider.Close()

	data := admin.Data{}
	data.Port = goapp.Config.GetInt("port")
	keysManager, err := mongodb.NewKeySaver(mongoSessionProvider, goapp.Config.GetInt("keySize"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't init saver"))
	}
	data.KeyGetter, data.KeySaver, data.OneKeyUpdater = keysManager, keysManager, keysManager
	data.OneKeyGetter = keysManager

	logManager, err := mongodb.NewLogGetter(mongoSessionProvider)
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't init log saver"))
	}
	data.LogGetter = logManager

	prStr := goapp.Config.GetString("projects")
	goapp.Log.Infof("Projects: %s", prStr)
	pv, err := admin.NewProjectConfigValidator(prStr)
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't init project validator"))
	}
	data.ProjectValidator = pv
	if err := mongoSessionProvider.CheckIndexes(pv.Projects()); err != nil {
		log.Fatal(errors.Wrap(err, "can't check indexes"))
	}

	printBanner()

	err = admin.StartWebServer(&data)
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't start the service"))
	}
}

var (
	version string
)

func printBanner() {
	banner := `
     ___    ____  ____                   __            __       
    /   |  / __ \/  _/        ____ _____/ /___ ___     \ \      
   / /| | / /_/ // /   ______/ __ ` + "`" + `/ __  / __ ` + "`" + `__ \_____\ \     
  / ___ |/ ____// /   /_____/ /_/ / /_/ / / / / / /_____/ /     
 /_/  |_/_/   /___/         \__,_/\__,_/_/ /_/ /_/     /_/  
  __               __                                     
 / /          ____/ /___  ____  _________ ___  ____ _____ 
/ / ______   / __  / __ \/ __ \/ ___/ __ ` + "`" + `__ \/ __ ` + "`" + `/ __ \
\ \/_____/  / /_/ / /_/ / /_/ / /  / / / / / / /_/ / / / /
 \_\        \__,_/\____/\____/_/  /_/ /_/ /_/\__,_/_/ /_/  v: %s

%s
________________________________________________________                                                 

`
	cl := color.New()
	cl.Printf(banner, cl.Red(version), cl.Green("https://github.com/airenas/api-doorman"))
}
