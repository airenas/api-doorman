package main

import (
	"strings"

	"github.com/airenas/api-doorman/internal/pkg/utils"

	"github.com/airenas/api-doorman/internal/pkg/mongodb"
	"github.com/airenas/api-doorman/internal/pkg/service"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/labstack/gommon/color"
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
	data.Handlers, err = initFromConfig(goapp.Sub(goapp.Config, "proxy"), mongoSessionProvider)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't init handlers"))
	}
	data.Port = goapp.Config.GetInt("port")

	utils.DefaultIPExtractor, err = utils.NewIPExtractor(goapp.Config.GetString("ipExtractType"))
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't init IP extractor"))
	}

	printBanner()

	err = service.StartWebServer(&data)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't start the service"))
	}
}

func initFromConfig(cfg *viper.Viper, ms *mongodb.SessionProvider) ([]service.HandlerWrap, error) {
	res := make([]service.HandlerWrap, 0)
	if cfg == nil {
		return nil, errors.New("Can't init handlers - names are not provided")
	}
	strHand := cfg.GetString("handlers")
	for _, sh := range strings.Split(strHand, ",") {
		sh = strings.TrimSpace(sh)
		if sh != "" {
			h, err := service.NewHandler(sh, cfg, ms)
			if err != nil {
				return nil, errors.Wrapf(err, "Can't init handler '%s'", sh)
			}
			res = append(res, h)
		}
	}
	return res, nil
}

var (
	version string
)

func printBanner() {
	banner := `
     ___    ____  ____                             __       
    /   |  / __ \/  _/                             \ \      
   / /| | / /_/ // /   _____________________________\ \     
  / ___ |/ ____// /   /_____/_____/_____/_____/_____/ /     
 /_/  |_/_/   /___/                                /_/      
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
