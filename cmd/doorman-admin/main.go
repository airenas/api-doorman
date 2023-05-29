package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/admin"
	"github.com/airenas/api-doorman/internal/pkg/integration/cms"
	"github.com/airenas/api-doorman/internal/pkg/mongodb"
	"github.com/airenas/api-doorman/internal/pkg/reset"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/labstack/gommon/color"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func main() {
	goapp.StartWithDefault()
	ctx := context.Background()

	config := goapp.Config
	mongoSessionProvider, err := mongodb.NewSessionProvider(config.GetString("mongo.url"))
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
	data.OneKeyGetter, data.UsageRestorer = keysManager, keysManager

	logManager, err := mongodb.NewLogProvider(mongoSessionProvider)
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

	data.CmsData = &cms.Data{}
	data.CmsData.ProjectValidator = pv
	data.CmsData.Integrator, err = mongodb.NewCmsIntegrator(mongoSessionProvider, goapp.Config.GetInt("keySize"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't init integrator"))
	}

	printBanner()

	tData := reset.TimerData{}
	tData.Reseter, err = mongodb.NewReseter(mongoSessionProvider)
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't init reseter"))
	}
	tData.Projects, err = initProjectReset(pv.Projects(), config)
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't init project rest config"))
	}
	data.UsageReseter = tData.Reseter

	ctxTimer, cancelFunc := context.WithCancel(ctx)
	doneCh, err := reset.StartTimer(ctxTimer, &tData)
	if err != nil {
		goapp.Log.Fatalf("can't start timer: %v", err)
	}

	err = admin.StartWebServer(&data)
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't start the service"))
	}
	cancelFunc()
	select {
	case <-doneCh:
		goapp.Log.Info("All code returned. Now exit. Bye")
	case <-time.After(time.Second * 15):
		goapp.Log.Warn("Timeout gracefull shutdown")
	}
}

func initProjectReset(projects []string, config *viper.Viper) (map[string]float64, error) {
	res := map[string]float64{}
	for _, p := range projects {
		v := config.GetFloat64(fmt.Sprintf("%s.MonthlyReset", p))
		if v > 0 {
			res[p] = v
		}
	}
	return res, nil
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
