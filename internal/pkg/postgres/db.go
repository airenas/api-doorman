package postgres

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// NewKeyValidator creates KeyValidator instance
func NewDB(ctx context.Context, dsn string) (*sqlx.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty DSN")
	}
	log.Ctx(ctx).Debug().Msg("Connecting to DB")
	sqlxDB, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to DB: %w", err)
	}
	sqlxDB.SetMaxOpenConns(10)
	sqlxDB.SetMaxIdleConns(10)
	if err := sqlxDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping DB: %w", err)
	}
	log.Ctx(ctx).Debug().Msg("Connected to DB")
	return sqlxDB, nil
}
