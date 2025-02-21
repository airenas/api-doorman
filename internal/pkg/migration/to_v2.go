package migration

import (
	"context"

	"github.com/airenas/api-doorman/internal/pkg/migration/api"
	"github.com/airenas/api-doorman/internal/pkg/postgres"
	"github.com/rs/zerolog/log"
)

type (
	Data struct {
		Project string
		Key     string
		AdmRepo *postgres.AdminRepository
		CmsRepo *postgres.CMSRepository
		DryRun  bool
	}
)

func Migrate(ctx context.Context, srvData *Data, in []*api.Key) error {
	user, err := srvData.AdmRepo.ValidateToken(ctx, srvData.Key, "")
	if err != nil {
		return err
	}
	log.Info().Str("user", user.ID).Str("name", user.Name).Msg("logged")

	return srvData.CmsRepo.CreatePlain(ctx, user, in, srvData.Project, srvData.DryRun)
}
