package main

import (
	"context"
	"fmt"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/admin"
	"github.com/airenas/api-doorman/internal/pkg/integration/cms"
	"github.com/airenas/api-doorman/internal/pkg/postgres"
	"github.com/airenas/api-doorman/internal/pkg/reset"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/labstack/gommon/color"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func main() {
	goapp.StartWithDefault()
	log.Logger = goapp.Log
	zerolog.DefaultContextLogger = &goapp.Log

	ctx := context.Background()
	err := mainInt(ctx)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}

func mainInt(ctx context.Context) error {
	config := goapp.Config

	db, err := postgres.NewDB(ctx, goapp.Config.GetString("db.dsn"))
	if err != nil {
		return fmt.Errorf("init db: %w", err)
	}
	defer db.Close()

	repo, err := postgres.NewAdmimRepository(ctx, db)
	if err != nil {
		return fmt.Errorf("init repository: %w", err)
	}

	data := admin.Data{}
	data.Port = goapp.Config.GetInt("port")
	data.KeyGetter, data.KeySaver, data.OneKeyUpdater = repo, repo, repo
	data.OneKeyGetter, data.UsageRestorer = repo, repo
	data.LogProvider = repo

	prStr := goapp.Config.GetString("projects")
	log.Info().Msgf("Projects: %s", prStr)
	pv, err := admin.NewProjectConfigValidator(prStr)
	if err != nil {
		return fmt.Errorf("init project validator: %w", err)
	}
	data.ProjectValidator = pv

	data.CmsData = &cms.Data{}
	data.CmsData.ProjectValidator = pv

	hasher, err := utils.NewHasher(goapp.Config.GetString("hashSalt"))
	if err != nil {
		return fmt.Errorf("init hasher: %w", err)
	}

	cms, err := postgres.NewCMSRepository(ctx, db, goapp.Config.GetInt("keySize"), hasher)
	if err != nil {
		return fmt.Errorf("init integrator: %w", err)
	}
	data.CmsData.Integrator = cms

	printBanner()

	tData := reset.TimerData{}
	tData.Reseter = repo
	tData.Projects, err = initProjectReset(pv.Projects(), config)
	if err != nil {
		return fmt.Errorf("init project rest config: %w", err)
	}
	data.UsageReseter = tData.Reseter

	ctxTimer, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()
	doneCh, err := reset.StartTimer(ctxTimer, &tData)
	if err != nil {
		return fmt.Errorf("start timer: %w", err)
	}

	err = admin.StartWebServer(&data)
	if err != nil {
		return fmt.Errorf("start web server: %w", err)
	}
	cancelFunc()
	select {
	case <-doneCh:
		log.Info().Msg("All code returned. Now exit. Bye")
	case <-time.After(time.Second * 15):
		log.Warn().Msg("Timeout graceful shutdown")
	}
	return nil
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
