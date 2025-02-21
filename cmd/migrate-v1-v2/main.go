package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/airenas/api-doorman/internal/pkg/admin"
	"github.com/airenas/api-doorman/internal/pkg/migration"
	"github.com/airenas/api-doorman/internal/pkg/migration/api"
	"github.com/airenas/api-doorman/internal/pkg/postgres"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/labstack/gommon/color"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	db, err := postgres.NewDB(ctx, config.GetString("db.dsn"))
	if err != nil {
		return fmt.Errorf("init db: %w", err)
	}
	defer db.Close()

	hasher, err := utils.NewHasher(config.GetString("hashSalt"))
	if err != nil {
		return fmt.Errorf("init hasher: %w", err)
	}

	mData := migration.Data{}

	mData.AdmRepo, err = postgres.NewAdmimRepository(ctx, db, hasher)
	if err != nil {
		return fmt.Errorf("init repository: %w", err)
	}

	prStr := config.GetString("projects")
	log.Info().Msgf("Projects: %s", prStr)
	pv, err := admin.NewProjectConfigValidator(prStr)
	if err != nil {
		return fmt.Errorf("init project validator: %w", err)
	}
	mData.Project = config.GetString("project")
	mData.Key = config.GetString("key")
	mData.DryRun = !config.GetBool("commit")
	log.Info().Str("project", mData.Project).Msg("Project")

	mData.CmsRepo, err = postgres.NewCMSRepository(ctx, db, config.GetInt("keySize"), hasher)
	if err != nil {
		return fmt.Errorf("init integrator: %w", err)
	}

	printBanner()

	log.Info().Msg("Migration started")

	if pv.Check(mData.Project) {
		log.Info().Msg("Project is valid")
	} else {
		return fmt.Errorf("project '%s' is not valid: allowed %v", mData.Project, pv.Projects())
	}

	log.Info().Msg("Reading data")

	var data []*api.Key
	if err := json.NewDecoder(os.Stdin).Decode(&data); err != nil {
		return fmt.Errorf("read data: %w", err)
	}

	log.Info().Int("len", len(data)).Msg("Data read")
	if err := migration.Migrate(ctx, &mData, data); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	log.Info().Msg("Migration finished")

	return nil
}

var (
	version string = "dev"
)

func printBanner() {
	banner := `

       __                                     
  ____/ /___  ____  _________ ___  ____ _____ 
 / __  / __ \/ __ \/ ___/ __ ` + "`" + `__ \/ __ ` + "`" + `/ __ \
/ /_/ / /_/ / /_/ / /  / / / / / / /_/ / / / /
\__,_/\____/\____/_/  /_/ /_/ /_/\__,_/_/ /_/ 
                                              
              _                  __  _                     ___     __            ___ 
   ____ ___  (_)___ __________ _/ /_(_)___  ____     _   _<  /     \ \     _   _|__ \
  / __ ` + "`" + `__ \/ / __ ` + "`" + `/ ___/ __ ` + "`" + `/ __/ / __ \/ __ \   | | / / /  _____\ \   | | / /_/ /
 / / / / / / / /_/ / /  / /_/ / /_/ / /_/ / / / /   | |/ / /  /_____/ /   | |/ / __/ 
/_/ /_/ /_/_/\__, /_/   \__,_/\__/_/\____/_/ /_/    |___/_/        /_/    |___/____/ 
            /____/ v: %s

%s
________________________________________________________                                                 

`
	cl := color.New()
	cl.Printf(banner, cl.Red(version), cl.Green("https://github.com/airenas/api-doorman"))
}
