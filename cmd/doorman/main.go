package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/airenas/api-doorman/internal/pkg/postgres"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/airenas/api-doorman/internal/pkg/service"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/labstack/gommon/color"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func main() {
	goapp.StartWithDefault()
	log.Logger = goapp.Log
	zerolog.DefaultContextLogger = &goapp.Log

	if err := mainInt(context.Background()); err != nil {
		log.Fatal().Err(err).Send()
	}
}

func mainInt(ctx context.Context) error {
	db, err := postgres.NewDB(ctx, goapp.Config.GetString("db.dsn"))
	if err != nil {
		return fmt.Errorf("init db: %w", err)
	}
	defer db.Close()

	hd := &service.HandlerData{DB: db}
	hd.Hasher, err = utils.NewHasher(goapp.Config.GetString("hashSalt"))
	if err != nil {
		return fmt.Errorf("init hasher: %w", err)
	}

	data := service.Data{}
	data.Handlers, err = initFromConfig(goapp.Sub(goapp.Config, "proxy"), hd)
	if err != nil {
		return fmt.Errorf("init handlers: %w", err)
	}
	data.Port = goapp.Config.GetInt("port")

	utils.DefaultIPExtractor, err = utils.NewIPExtractor(goapp.Config.GetString("ipExtractType"))
	if err != nil {
		return fmt.Errorf("init IP extractor: %w", err)
	}

	printBanner()

	err = service.StartWebServer(&data)
	if err != nil {
		return fmt.Errorf("start web server: %w", err)
	}
	return nil
}

func initFromConfig(cfg *viper.Viper, hd *service.HandlerData) ([]service.HandlerWrap, error) {
	res := make([]service.HandlerWrap, 0)
	if cfg == nil {
		return nil, errors.New("Can't init handlers - names are not provided")
	}
	strHand := cfg.GetString("handlers")
	for _, sh := range strings.Split(strHand, ",") {
		sh = strings.TrimSpace(sh)
		if sh != "" {
			h, err := service.NewHandler(sh, cfg, hd)
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
