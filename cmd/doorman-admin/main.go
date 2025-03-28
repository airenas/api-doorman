package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/admin"
	"github.com/airenas/api-doorman/internal/pkg/handler"
	"github.com/airenas/api-doorman/internal/pkg/integration/cms"
	"github.com/airenas/api-doorman/internal/pkg/model"
	"github.com/airenas/api-doorman/internal/pkg/model/permission"
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

	hasher, err := utils.NewHasher(goapp.Config.GetString("hashSalt"))
	if err != nil {
		return fmt.Errorf("init hasher: %w", err)
	}

	repo, err := postgres.NewAdmimRepository(ctx, db, hasher)
	if err != nil {
		return fmt.Errorf("init repository: %w", err)
	}

	data := admin.Data{}
	data.Hasher = hasher
	data.Port = goapp.Config.GetInt("port")
	data.UsageRestorer = repo
	data.OneKeyGetter, data.LogProvider = repo, repo
	authmw, err := handler.NewAuthMiddleware(repo)
	if err != nil {
		return fmt.Errorf("init auth middleware: %w", err)
	}
	data.Auth = authmw.Handle

	prStr := goapp.Config.GetString("projects")
	log.Info().Msgf("Projects: %s", prStr)
	pv, err := admin.NewProjectConfigValidator(prStr)
	if err != nil {
		return fmt.Errorf("init project validator: %w", err)
	}
	data.ProjectValidator = pv

	data.CmsData = &cms.Data{}
	data.CmsData.ProjectValidator = pv

	cms, err := postgres.NewCMSRepository(ctx, db, goapp.Config.GetInt("keySize"), hasher)
	if err != nil {
		return fmt.Errorf("init integrator: %w", err)
	}
	data.CmsData.Integrator = cms

	utils.DefaultIPExtractor, err = utils.NewIPExtractor(goapp.Config.GetString("ipExtractType"))
	if err != nil {
		return fmt.Errorf("init IP extractor: %w", err)
	}

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

	if err := tryAddInitialAdmin(ctx, config, repo, pv.Projects()); err != nil {
		return err
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

func tryAddInitialAdmin(ctx context.Context, config *viper.Viper, repo *postgres.AdminRepository, projects []string) error {
	key := config.GetString("mainAdmin.key")
	if key == "" {
		return nil
	}

	if len(key) < 30 && !config.GetBool("mainAdmin.forceShortKey") {
		return fmt.Errorf("admin auth key too short")
	}

	log.Ctx(ctx).Info().Msg("Try to add initial admin")
	err := repo.AddAdmin(ctx, key, &model.User{Name: "Main",
		MaxLimit:    config.GetFloat64("mainAdmin.maxLimit"),
		MaxValidTo:  time.Now().AddDate(10, 0, 0),
		Permissions: map[permission.Enum]bool{permission.Everything: true},
		Projects:    projects})
	if err != nil {
		if errors.Is(err, model.ErrDuplicate) {
			log.Ctx(ctx).Info().Msg("Initial admin exists")
			return nil
		}
		return fmt.Errorf("add initial admin: %w", err)
	}
	log.Ctx(ctx).Warn().Msg("Initial admin added")
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
